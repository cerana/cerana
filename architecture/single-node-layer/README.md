# The Node
A node is something running mistify-os. In the real world, this probably amounts to either to real hardware or a virtual machine. The node has one simple objective: Join a cluster of other nodes on the same network fabric.

## Booting
With the exception of [bootstrapping](../bootstrapping), booting is designed to be simple: Using pxeboot, we get everything we need to setup and start the box. If it is a first time boot, there may be some configuration input (see [boot-process](../boot-process), otherwise the box will come up.
