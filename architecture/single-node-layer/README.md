# The Node
A node is something running mistify-os. In the real world, this probably amounts to either to real hardware or a virtual machine. The node has one simple objective: Join a cluster of other nodes on the same network fabric.

## The kernel and modules
Mistify-os uses (generally) the latest linux kernel and modules. We add things we think make for a better system:
* ZFS on Linux project (zfs module)
* Kernel patches for zfs delegation via namespaces

In general, for images and volumes, use of zfs is encouraged (and for official mistify projects, required).

## Booting
With the exception of [bootstrapping](../bootstrapping/README.md), booting is designed to be simple: Using pxeboot, we get everything we need to setup and start the box. If it is a first time boot, there may be some configuration input (see [boot-process](../boot-process/README.md)), otherwise the box will come up.

## Components of a node
A single node needs 4 basic things to function as a good member of the cluster:
* [Process management](../process-management)
* [Distributed key/value store](../distributed-key-value-store)
* [Distributed DHCP server](../distributed-dhcp)
* [Task subsystem](../task-subsystem)

Anything in a mistify-os image should either contribute to booting up, or be a component to one of these 4 things. Mistify-os will ship the task-coordinator and some basic task-providers as part of the image, but has means (through tasks) to add additional task-providers.

## Where are the tools to do (x)?

Mistify-os ships with a very capable kernel, with additional modifications added by this project. That being said, the tools to take advantage of it (example: ebpf userland stuff for debugging) will be missing. This is simple: you can ship those tools (or anything) as an image to the box and run it. Mistify-os ships as small as possible, while striving to be extremely capable. The rest is up to an operator's imagination.
