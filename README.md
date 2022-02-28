# matching_engine
===============

A simple financial trading matching engine. Built to learn more about how they work.

This system is not a complete implementation of a full matching engine system. It contains a basic set of internal components for a matching engine but doesn't include any networking components which would be required to actually use it in production.

## msg

The `msg` package defines the basic messages which are processed by a matching engine. The system was designed to be _entirely_ driven by these messages, so we find both the expected buy/sell limit-order type messages as well as new-trader and shutdown messages which change the state of the matching engine without actually engaging it trading activity.

The rationale behind this choice was that the state of a matching engine should be determined _exclusively_ by the serise of messages which have passed through it. This allows us to perfectly reproduce any matching engine state simply by replaying these messages.

There is some initial work done on a binary serialisation format. This was intended to be used for storing messages in logs, and for sending over the network. I don't think the approach taken here was very well conceived, and I would probably use an efficient encoding library if I was to work on this now.

The `maker.go` file is interesting in that it generates random sets of buy/sell messages. While the code-base has many traditional run-code/assert-outcome style unit tests, we also employed a lot of system-invariant style tests with randomly generated test data.

## matcher/pqueue

This is a priority queue implementation custom built to support the matching engine. This is definitely the most complex piece of code in the system. The pqueue is a pair of linked red-black trees. The two trees share the nodes. Specifically an `OrderNode` is a member of two red-black trees, one tree is ordered by price and the other is ordered by the id of the order. A feature of having a single order object linked to two trees is that when we remove an `OrderNode` from one tree we can remove it from the other, without having to perform a search through the tree to find it.

The red-black tree is an internal detail the publicly exposed type is the `pqueue.MatchQueues` this is a trio of red-black trees, one for buys and one for sells and one for all orders sorted by guid. This exposes the operations needed to submit a buy, or sell order as well as cancel an existing order (using its guid).

The testing approach used here is primarily invariant based testing. We perform operations on the rb-trees and then test to ensure that the structural invariants of the trees has not been violated. In the development of these trees I first started with a body of hand-written operations, followed with assertions on the result. Even after writing a very large number of unit tests I was not able to find any bugs. However, when I built the randomly generated invariant based testing I was able to find a number of subtle bugs and fix them

If I did this now I would have kept the 'normal' style unit tests as well as the more complex invariant style tests. Although invariant testing was more valuable in finding bugs, the 'normal' style of unit tests are easy to read and are a useful form of documentation of the expected behaviour of the rb-trees.

## matcher

The matcher implements an actual matching engine. This uses a `pqueue.MatchQueues` to manage incoming orders. As each new order comes in an attempt is made to match the order, buy or sell, and the resulting matches are written to the output. Cancelling orders is supported, as-is shutting down the order book.

## coordinator

This package is designed to allow us to wrap a `matcher.M` with an input and output queue. There are two implementations available, one which uses a Go channel and one which uses an imported high performance queue. The queue imported is from another project I authored which can be found at `github.com/fmstephe/flib`.

I would not use this approach if I was building this system again today. I think that the choice to make the `matcher.M` struct embed the `coordinator.AppMsgHelper` interface is unnecessarily complicated.
