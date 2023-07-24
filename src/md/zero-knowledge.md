title: Zero Knowledge about ZK
date: 2023-07-24
---

In the recent months, I've been diving into ZK and cryptography. I will be using this space to write personal notes on protocols and algorithms I learn about, as a way to reinforce my own knowledge.

This will serve as my introductory post.

## What is a Zero Knowledge proof?

Simply put, it is a receipt that proves that you (prover) know a certain statement, and this receipt is usually given to another person (verifier) to accept or reject.

Different people use different examples to illustrate this, but I personally found [@cryptographor](https://twitter.com/cryptograthor)'s example the best: [the color blindness test](https://youtu.be/rkFrGz-JGzo?t=163).

## Color blindness test

Imagine that Alice is color blind and Bob wants to prove to Alice that he can see color. Now imagine that Alice has a red ball and a blue ball. Since she is color blind, she is not convinced that the balls have color, and Bob wants to convince her of this fact.

A simple test that Bob can do to prove that he can see color would be to tell Alice to hold each ball in her hands behind her back and then either swap or not swap the balls in her hands before revealing them to Bob. Bob then needs to tell Alice whether he thinks Alice swapped the balls or not.

In the first iteration, Alice might think that Bob got lucky, since there was a 50% chance that Bob guessed correctly, but after enough iterations of the protocol above, Alice can be convinced that Bob is telling the truth if he is consistently correct over all the iterations.

## Properties of ZK

The above example also illustrates the 3 properties of ZK: **completeness**, **soundness**, **zero-knowledge**:

- **Completeness**
Alice is convinced by an honest prover that color exists and that Bob can tell the difference
- **Soundness**
If Bob is lying about the existence of color, it should be near impossible to convince Alice of that falsehood.
- **Zero-Knowledge**:
Alice learns nothing other than the statement is true (color exists).

## Modelling ZK statements

Programs do not understand English statements like the above, so we want to convert our English statement to some sort of verifiable statement. This is usually done in the form of some mathematical computation, which comes in the form of **circuits**.

Circuits allow us to express boolean (`AND`, `OR`) and arithmetic (`+`, `*`) logic. We can then model our problem as a circuit of boolean and arithmetic gates which in turn allow us to express the  **constraints** of the program, which are pre-set restrictions placed on the circuit. The circuit also usually takes in witness values from the prover as private input.

Once the circuit is formed and the input is provided, the circuit can be evaluated and a proof can be obtained. The verifier can be convinced that the proof is valid if all constraints within the circuit are satisfied - in plain English, this means that the prover has shown the verifier that he knows a given statement.
