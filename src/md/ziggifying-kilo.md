title: Ziggifying Kilo
date: 2023-04-13
---

Recently, I ported antirez's [Kilo editor](https://github.com/antirez/kilo) to Zig. It's called *gram* (*because kilogram, duh*) and I tried to keep
it as faithful as possible to the original:

- *Under 1000 LOC*. `cloc` on `main.zig` alone yields *734* lines of Zig (versus the *986* in the original)
- *Has almost the same features* (except non-prints support)
- Written as similarly as I could to the original but with Zig fixes to the C-*isms*.

There are already some comparisons of Rust vs Zig in the wild (*I quite enjoyed 
[When Zig is safer and faster than Rust](https://zackoverflow.dev/writing/unsafe-rust-vs-zig) 
by [@zack_overflow](https://twitter.com/zack_overflow)*) - this post will be on Zig vs C from the eyes
of someone new to low-level programming.

## The Road to Zig 1.0

In [The Road to Zig 1.0](https://www.youtube.com/watch?v=Gv2I7qTux7g), Andrew Kelley, the creator of Zig,
describes Zig as "C but with the problems fixed".

He lists 3 major problems with C:

1) `#include` - causes slow compilations and prevents optimizations

2) other preprocessor macros like `#define` - makes it harder to read and debug

3) undefined behaviour footguns everywhere

The first point is irrelevant since I'm writing a tiny text editor, but while porting over kilo to *gram*,
I often encountered the above problems while reading the kilo source code, and to my pleasant surprise
the Zig version felt a lot cleaner to read by the end. I'll evaluate my experience based on how well I think
Zig fixes problems 2) and 3). 

## Preprocessors begone!

The entire *gram* editor is written in a single `main.zig` file and `std` is the only dependency. 

No `#include` and `#define` shenanigans - in Zig, everything is just a `const`:

```zig
// C: #define KILO_QUIT_TIMES 3
const GRAM_QUIT_TIMES = 3;
```

While macros are simply functions: 

```zig
// In C:
// #define FIND_RESTORE_HL do { \
//     if (saved_hl) { \
//         memcpy(E.row[saved_hl_line].hl,saved_hl, E.row[saved_hl_line].rsize); \
//         free(saved_hl); \
//         saved_hl = NULL; \
//     } \
// } while (0)

// In Zig:
fn findRestoreHighlight(
    self: *Self,
    saved_hl: *?[]Highlight,
    saved_hl_ix: ?usize,
) void {
    if (saved_hl.*) |hl| {
        mem.copy(Highlight, self.rows.items[saved_hl_ix.?].hl, hl);
        saved_hl.* = null;
    }
}
```

In *gram*, highlight related definitions are simply a Zig `enum` instead of a bunch of
`#define` directives, which allows the usage of the handy `@enumToInt` to derive syntax color instead:

```zig
const Highlight = enum(u8) {
    number = 31,
    match = 34,
    string = 35,
    comment = 36,
    normal = 37,
};

        ...
        var color = @enumToInt(hl);
        ...
```

The above are just some of the many examples where Zig has made the code far more human-readable without
sacrificing the conciseness of C.

## UB footguns?

> The simple, lazy way to write code must perform robust error handling.

UB is bad, but unnecessary UB is worse - this is something that I think Zig remedies with the way you are forced
to write programs.

Consider the [rewritten save functionality](https://github.com/bingcicle/gram/blob/0b79b81b539bcf349012f2ea1ff862854b707dd7/src/main.zig#L543) in *gram*:

```zig
fn save(self: *Self) !void {
    const buf = try self.rowsToString();
    defer self.allocator.free(buf);

    const file = try fs.cwd().createFile(
        self.file_path,
        .{
            .read = true,
        },
    );
    defer file.close();

    file.writeAll(buf) catch |err| {
        return err;
    };

    try self.setStatusMessage("{d} bytes written on disk", .{buf.len});
    self.dirty = false;
    return;
}
```

In the original kilo's save functionality alone, there already exists a bunch of indirection with
 a `writeerr` goto which is referenced a total of 3 times to handle the same error,
and the failure case where an error message is written to the status message is also handled
in the same function.

Contrast this with *gram* above, where it's a pretty clear linear flow with coupled resource creation/cleanup
and error handling. Setting the status message within this function only happens if this
succeeds, otherwise we simply catch and return the error in order to set the status message higher
up.

The best way to illustrate Zig's effect on my way of thinking is that it 
*quietly and gently nudges you to think about where data lives in terms of allocation, cleanup
and error handling*.

The simplicity and linearity of Zig seemed like a con to me at first, but the mental model 
that Zig forces me into is refreshing and so far I'm enjoying the ride.

Feel free to reach out on [my twitter](https://twitter.com/bingcicle) to give me feedback :)
