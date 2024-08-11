# smegmesh

## Overview

Distributed WireGuard mesh management. This tool helps to configure WireGuard
networks in a mesh topology such that there is no single point of failure.
The tool aims to set-up mesh networks with minimal knowledge and
configuration of WireGuard.

The idea being that a node can take up one of two roles in the network, a
peer or a client. A peer is publicly accessible and must have IPv6 forwarding
enabled. Peer's responsibility is routing traffic on behalf of clients
associated with it.

Whereas, a client hides behind a private endpoint in which all packets are
routed through the peer. A client must enable the flat `keepAliveWg` to
ensure that its associated peer learns about any NAT mappings that change.

IPv6 is used in the overlay to make use of the larger address space.
A node hashes it's WireGuard public key to create an identifier
(the last 64-bits of the IPv6 address) and the mesh-id is hashed into
the first 64-bits of the IPv6 address to create the locator.

A node (both client and a peer) can be in multiple meshes at the same
time. In which case the node can optionally choose to act as a bridge
and forward packets between the two meshes. Through this it is possible
to define complex topologies. To route between meshes multiple hops away
a simple link-state protocol is adopted (similar to RIP) in which the
path length (number of meshes) is used to determine the shortest path.

Redundant routing is possible to create multiple exit points to the same
mesh network. In which case consistent hashing is performed to split traffic
between the exit points.

## Message Dissemination

A variant of the gossip protocol is used for message dissemination. Each peer
in the network is ordered lexicographically ordered by their public key.
The node with the lexicographically lowest public key is used as the leader
of the mesh. Every `heartBeatInterval` disseminates a refresh message
throughout the entirety of the group in order to prune nodes that may
have prematurely died.

If after `3 * heartBeatInterval` a node has not received a dissemination
message then the node prunes the leader and expects one from the next
lexicographically lowest public key.

To 'merge' updates and reconcile any conflicts a Conflict Free Replicated
Data Type (CRDT) is implemented. Consisting of an add and remove set.
Where a node is in the group if it is in the add set and there is either
no entry in the remove set or the timestamp in the remove set has a lower
vector clock value.

## Performance

This prototype has been tested to a scale of 3000 peers in the network.
Furthermore, the fault-tolerance has been tested to a scale 3000 nodes
to the order of 20 seconds for the entire network and 12 seconds
for the 99 percentile.

## Installation

To build the project do: `go build -v ./...`. A Docker file is provided
to get started.

To build with the Dockerfile:
`docker build -t smegmesh-base ./`

Then run an example topology in the examples folder. For example:
`cd examples/simple && docker-compose up -d`

## Tools

### Smegd
Smegmesh requires the daemon process to be running (smegd) which also takes
a configuration.yaml file. An example yaml configuration file is provided in
examples/simple/shared/configuration.

### Smegctl
Smegctl is a CLI tool to create, join, visualise and administer networks.

### Api
An api is provided to invoke functions to create, join, visualise and administer
networks. This could be used to create an application that allows a user
to configure the networks.

### Dns
A dns server is provided to resolve an alias into an IPv6 address.

