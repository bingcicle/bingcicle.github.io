title: Grug Brained Guide to KZG Polynomial Commitments
date: 2024-01-07
---

In this post, I'm aiming to explain to myself what a KZG polynomial commitment is,
and talk about my [Zig implementation](https://github.com/bingcicle/kzigg).
If you're only curious about the Zig usage, feel free to skip to the last section.

# Why another explainer?

Most blog posts/explainers I found out there start off with the heavy math or 
the deep technical components of what a KZG polynomial commitment is
right away. They also are often disconnected, going from one concept to the
next without making it very clear how it all links together. The real kicker
imo is that none really starts off with the motivation behind KZG in the
context of Ethereum and a brief technical overview first.
The closest I found was [@protolambda](https://twitter.com/protolambda)'s
[implementor notes](https://hackmd.io/@protolambda/eip-4844-implementer-notes).

So this post is more of a post to myself prior to learning about KZG, and is
meant to be a bridge to the more
[technical](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html)
[and](https://alinush.github.io/2020/05/06/kzg-polynomial-commitments.html)
[instructional](https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/)
[articles](https://scroll.io/blog/kzg) (each word here is a link!) that
already exists out there.

This post might be helpful if you think like me and require some motivating examples 
and a high level overview first to put the technical knowledge into practice.
Otherwise, more technical readers should probably read the above links instead.

# Why KZG commitments?

Before we understand what a KZG commitment is, I'd like to think that 
we should understand what they're used for first.

For a long while, Ethereum transaction fees have been way too expensive
for regular users. Every transaction you made had to be processed by
every participating validator in the network, and the fees that you pay is
the cost of processing those transactions.

Rollups were meant to solve this problem by bundling transactions together
but even rollup fees can get too expensive for many users since rollups 
still need to pay for the data posted onto mainnet,
and this posting of data is a function of how large the data
blobs are and the current L1 gas price.

Knowing the above, the solution is probably to

1) either post less data or not post the data at all, and/or
2) have some sort of gas-agnostic way to post the data

The long term solution is to shard data which takes time to implement,
so a stopgap solution is necessary to make fees cheaper for now. This stopgap
is [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844).
The crux of it is to introduce the *transaction format* that
will be used in sharding but not actually shard those transactions. This
(kinda) addresses the 2 problems above.
Notably, point 1 is where **KZG commitments** come in.

**"*either post less data or not post the data at all*"**

Currently, transaction data is stored within the calldata, which is visible
to the EVM and is a permanent part of the blockchain. EIP-4844 introduces 
blob-carrying transactions which makes Ethereum store data on the 
consensus layer rather than the execution layer (EVM). Rather, what the 
execution layer sees is the *commitment* to those blobs. These commitments
are smaller in size which saves on gas, and it is sufficient to verify
these commitments without needing to access the actual blobs.

Note here also that the consensus layer also does not store the blobs in perpetuity -
the role of the consensus layer is to provide a **secure** real-time view of what
is being published.

**"*have some sort of gas-agnostic way to post the data*"**

So 4844 solves this by introducing an entirely separate fee market for blobs
(which is why I said *kinda* above). This is out of scope of this post, since
it has nothing to do with how KZG works. Instead, the [proto-danksharding FAQ](https://notes.ethereum.org/@vbuterin/proto_danksharding_faq#What-does-the-proto-danksharding-multidimensional-fee-market-look-like) has a section that goes in-depth into the fee market
structure.

# KZG commitments

Now we can get to how we achieve point 1 mentioned above. Commitment schemes
allow one to publish a value which binds one to a message without revealing it.
One can then open the commitment and reveal the committed message to a verifier
to be checked. Of course this only makes sense if the cost of committing is 
less than the cost of sending the entire message.

[KZG (Kate-Zaverucha-Goldberg) commitments](https://www.iacr.org/archive/asiacrypt2010/6477178/6477178.pdf)
are a class of the above scheme.

Some key characteristics that is useful for 4844:

- Constant size commitment/proof (48 bytes). This is especially useful for batching,
since the proof size is always constant regardless of the size of the blob.
- Verification is a single pairing check (constant time)

# How it works, ELI5

The scheme itself is stupidly simple from an engineering POV and I didn't realize
this myself prior to implementing it (The math is super complicated though). 
@protolambda explained it best in the notes linked at the top: you only need

1) a linear combination to compute a KZG commitment,
2) a single pairing verification to verify a KZG proof

Obviously there are way more details behind how the above steps happen
(serialization/deserialization of blobs, how pairings work, optimizations, etc.),
but the above 2 steps is really all that is happening in the scheme.

Again, I would highly recommend the other articles for the math but nevertheless
I will give a short overview here - note that this is in the context of blobs.
I will also (try to) include the little details I noticed in the code
that some articles I've read seem to have missed out on.

## Setup

Some commitment schemes use some secret value within its computation, and this
secret value is often obtained via something called a
[**trusted setup**](https://ceremony.ethereum.org).

Essentially this is a multiparty procedure where each party creates some secret
and runs a computation to mix it with the previous contributions.
Eventually, the final secret value will be used for the commitment scheme.
This secret value will be used to compute **all** group elements that
are available to the prover and the verifier.

The cool thing about this trusted setup is that it has a
"1-of-N" trust assumption, which means only a single participant is required
to be honest for the procedure to be secure. That means, unless you don't
trust yourself, trusted setups are generally OK to trust.

Some fun numbers: there were **141,416** contributors to the 
[KZG ceremony](https://ceremony.ethereum.org/), and
the ongoing (as of writing this)
[Penumbra summoning ceremony](https://summoning.penumbra.zone/)
already has about **11,799** contributors!

## Commit

As mentioned earlier, a commitment is simply a linear combination. A blob of bytes
has its data transformed into a polynomial and then a linear combination done
on its points, producing a serialized G1 point (48 bytes in size) which serves
as the commitment. This is further compressed into a versioned hash (32 bytes)
for forward compatibility. 

This process can be done naively (very slow) or via Pippenger's algorithm.

## Prove

Now we want to show that we know the original data in the blob, otherwise the
polynomial. In other words, we want to show that we know $p(z) = y$, where
y is the evaluation of the polynomial at some point $z$.

The simplest way to do that is if the prover sends the entire polynomial
to the verifier, but that would defeat the point of the commitment scheme - we
would ideally want cost savings when we go through the trouble of using such a
scheme!

Now what is this is point $z$? This is usually called a **challenge** - this is
a random field element that the verifier sends to the prover, allowing the
prover to evaluate the polynomial to prove its integrity.
This requires the prover and verifier to communicate directly with each other,
which isn't ideal. Instead, we rely on the Fiat-Shamir heuristic by letting the
prover and verifier agree on a format prior to the protocol to simulate this interaction.
Both the prover and verifier independently calculate the challenge, by hashing
a 'transcript' (the simulation) into a field element, which serves as our challenge.

Using this challenge to evaluate the polynomial, we get the **evaluation**
and the **evaluation proof** that attests to the fact that the polynomial was
correctly evaluated in the eyes of the prover.

The evaluation proof is actually the **quotient polynomial** $q(x)$ such that

$$
q(x) = \frac{p(x) - y}{x - z}
$$

To derive this: we know that if $z$ is a root of $p(x)$, then $p(z) = 0$, 
Using this property, we can actually take advantage of the fact that $p(x) - y$
is zero at $z$ to express a quotient polynomial $q(x)$.


## Verification

Verification is then done with a pairing check, which I treated as a black box
for this post and implementation because I don't know enough to make
sensible comments. Instead, I've linked to Vitalik's
[blog post](https://medium.com/@VitalikButerin/exploring-elliptic-curve-pairings-c73c1864e627)
on the topic.

## Batching

Interestingly enough, we can re-express the above quotient polynomial in order to
batch prove across a set of points:

$$
q(x) = \frac{p(x) - i(x)}{z(x)}
$$

Here, $i(x)$ is a polynomial (in Lagrange form) of a set of points to prove, and
$z(x)$ is the **zero** or **vanishing polynomial** that is the set of linear
factors that can divide $p(x) - i(x)$.

Verification is then just using the same pairing check, except on linear
combinations of group elements instead! You can see this in action within
the code, where linear combinations (naively computed for security
reasons) are done in the batch verification function, both in my port as well
as the original version.

# Zig implementation

I ported [c-kzg-4844](https://github.com/ethereum/c-kzg-4844) to Zig
just to learn what goes behind the scheme. I wrote
previously about [porting another C project](https://bingcicle.github.io/posts/ziggifying-kilo.html)
to Zig so for general impressions you can check that post out.
I'll post thoughts specifically comparing my implementation with the C version.

## C Interop

For the Zig implementation, like the C version, I relied on [`blst`](https://github.com/supranational/blst)
for the backend, which means I had a chance to finally try C interop
after using Zig for some time. Zig natively supports
[C ABIs](https://ziglearn.org/chapter-4/), making it possible for new programs
to be built leveraging C libraries.

Frankly, it surprised me with how easy it was to just build `blst` and use it as a
static library in Zig. Zig pointers also coerce nicely to C pointers, which was
super convenient to not have to do weird casts when calling `blst`.

Why bother with C? Well, when a C library is audited and battle-tested, it
may be better to just use it rather than reinvent the wheel, which is exactly
what Zig enables you to do easily.

## Defer is great

In the C implementation, there's heavy abuse of `goto`s in order to free data
whenever some computation finishes or when errors occur. This is where C
gets a bad rap, because the need to do manual memory management along with
copious amounts of indirection leads to bug-prone code.

Instead of doing all that, Zig encourages the `alloc-defer` pair pattern:

```zig
    const allocator = std.testing.allocator;
    // This allocates
    const cfg = try KZGTrustedSetupConfig.loadFromFile(
        allocator,
        "./src/trusted_setup.txt",
    );
    // This frees when out of scope
    defer cfg.deinit();
```

This means less cognitive load on thinking about where and when to free your memory.

## General cleanliness

Generally, the Zig version felt slightly cleaner to me and came out to about
~1100LOC (including tests). The [c_kzg_4844.c](https://github.com/ethereum/c-kzg-4844/blob/main/src/c_kzg_4844.c)
file alone came to about 1.5x of that.

# Conclusion

Implementing the 4844 variant of KZG for myself was a good practice for me to uncover what goes
behind the scenes during KZG, and its connection to EIP-4844. While this isn't really
made to be used as a SNARK, it is still really cool to see commitment schemes
used in practice.

Commitment schemes are like magic when applied in the right scenarios to save on
space and the work to be done.
