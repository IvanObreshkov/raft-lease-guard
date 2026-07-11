# ADR 001 — Initial Architecture Plan

## Status
Accepted, 2026-05-24

## Introduction
After looking at one of the best well known implementations of Raft, namely [etcd/raft](https://github.com/etcd-io/raft)
I saw that it follows the so called [Sans-IO](https://sans-io.readthedocs.io/how-to-sans-io.html) (French for “without I/O”) 
architecture pattern, meaning the implementation does not have any I/O operations in it, it's just a pure state machine, 
the I/O is left to the upper levels in the stack.

This pattern allows for better testability and makes the implementation easier to reason about, as written [here](https://sans-io.readthedocs.io/how-to-sans-io.html#why-write-i-o-free-protocol-implementations).
This makes it a good foundational pattern for my project, especially if I want to test my implementation using a [Deterministic simulation testing (DST)](https://antithesis.com/docs/resources/deterministic_simulation_testing/) approach.

## Alternatives considered
I also checked [HashiCorp's implementation of Raft](https://github.com/hashicorp/raft). I did not like their approach, as it has multiple goroutines and I/O operations coupled with the Raft core logic, which makes DST quite hard to implement.
I also considered a purely functional Raft core, where the state is immutable and each transition returns a new state value. But Go is an imperative language and has no language-level support for immutability, so a purely functional approach is not the best fit for Go.
Also, the Sans-I/O pattern along with DST should be enough to validate correctness for this project.

## Primary interface
Etcd's implementation of Raft, has one primary method, that drives the transitions of the state machine - [Step()](https://github.com/etcd-io/raft/blob/5002aff469116236cb296038b926edcab16d05a9/raft.go#L1089)
and the caller of this method is the main loop in the [node.run()](https://github.com/etcd-io/raft/blob/5002aff469116236cb296038b926edcab16d05a9/node.go#L389) goroutine.
On each iteration of the loop the [HasReady()]() function is called, which checks whether the state machine has made a transition. This is done because the [Ready struct](https://github.com/etcd-io/raft/blob/5002aff469116236cb296038b926edcab16d05a9/node.go#L52) batches multiple state machine transitions,
so that they could be stored via a single write operation to durable storage.
This is a great optimization, but I will not focus on it, at least for now, as I want to first implement all the core parts correctly. This project is also solely for learning purposes, and is not intended for Production use.

Instead, I decided to do something simpler, inspired by the interface of [h11](https://github.com/python-hyper/h11), a Sans-I/O implementation of the HTTP/1.1 protocol in Python.
h11 exposes its state machine through methods like [next_event()](https://github.com/python-hyper/h11/blob/62c5068c971579d61fa1b55373390e12f25fd856/h11/_connection.py#L438) (which advances internal state and returns the next parsed event) and [send(event)](https://github.com/python-hyper/h11/blob/62c5068c971579d61fa1b55373390e12f25fd856/h11/_connection.py#L508) (which advances state and returns bytes to write).

I am going to model my Raft implementation as a [Mealy machine](https://en.wikipedia.org/wiki/Mealy_machine). This is a Finite State Machine (FSM), whose outputs depend on both the inputs to the machine, and the current state of the machine. Here is what I mean:
```go
func (r *Raft) Transition(msg Msg) []Action
```
The Transition function, accepts a `Msg` as inputs (e.g, `AppendEntriesRequest`), and mutates the state of the Raft Node (e.g, `Follower` -> `Candidate`), based on the `Msg` and the current state of the Node. As outputs, it returns a list of Actions, that should be taken by the upper, layers (e.g, returning an `AppendEntriesResponse`). 
More formally:

$T: S \times \Sigma \rightarrow S \times \Lambda$, where

* $T$ is the transition function.
* $\Sigma$ is the inputs, a.k.a. the `Msg`.
* $S$ is the state of the Node - the set of Raft variables (role, currentTerm, log, etc.). It is not returned explicitly, but is mutated in place on the Raft struct.
* $\Lambda$ is the output, a.k.a. the Actions that should be taken by the upper layers.

## The upper layers
The `Transition` method is the sans-I/O protocol core.  I envision the upper layers to have input Go-channels which receive different Msgs (RPCs, Ticks, etc.), and they pass those to the `Transition` function. I would also have output channels, which forward the Actions to the relevant I/O consumers (e.g., Durable Storage, gRPC clients).

