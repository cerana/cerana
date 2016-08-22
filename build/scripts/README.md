Setting Up a Test Environment Using VMs
=======================================

This document describes scripts which can be used to setup a network configuration and start virtual machines for testing simple Cerana clusters on a Linux host. They support any number of virtual machines limited only by the capabilities of the host.

Each script maintains a configuration in `~/.testcerana`. This directory contains all of the options with with the scripts were last run making it unnecessary to remember the options when repeating the same test scenarios.

To reset an option to its default value simply use the value "default". e.g. To reset the disk image directory option (`start-vm`) use `--diskimagedir default`.

The following is an overview of the purpose and capabilities of each script. Each script also provides a comprehensive list of command line options which can be viewed using the `--help` option. Basically, the workflow is to use `vm-network` to configure the network for a test scenario and then use `start-vm` one or more times to start each of the VMs for the test.

cerana-functions.sh
-------------------

This script contains a number of helper functions used by the other scripts.

vm-network
----------

**NOTE:** Because this script reconfigures the network this script uses `sudo` to gain root access.

Testing interaction between the various nodes of a Cerana cluster requires a network configuration which allows communication between the nodes but avoids flooding the local network with test messages. This is accomplished using a bridge to connect a number of [TAP](https://en.wikipedia.org/wiki/TUN/TAP) devices. When using this script it helps to keep the following in mind. * A Cerana cluster can be comprised of 1 or more VMs. Each VM can have 1 or more TAP interfaces. The TAP interfaces to be associated with a specific VM is termed a "tapset". The number of *tapsets* is controlled using the option `--numsets`.

By default TAP interfaces are named using using the pattern `ceranatap.<tapset>.<n>` where `<n>` is TAP number within a *tapset*. For example a configuration having three VMs with two interfaces each produces three *tapsets* each having two interfaces. The resulting TAP interfaces become:

```
        ceranatap.1.1
        ceranatap.1.2
        ceranatap.2.1
        ceranatap.2.2
        ceranatap.3.1
        ceranatap.3.2
```

Each TAP interface is assigned a MAC address beginning with the default pattern "DE:AD:BE:EF". The 5th byte of the MAC address is the number of the corresponding *tapset* and the 6th byte is the TAP number within the *tapset*.

These are then all linked to a single bridge having the default name `ceranabr.1`.

**NOTE:** Currently only a single bridge is created. In the future multiple bridges will be used to better support testing VMs having multiple TAP interfaces. For example one bridge can be used for the node management interfaces while a second bridge can be used for a connection to a wider network.

The `vm-network` script also supports maintaining multiple network configurations making it easy to tear down one configuration and then setup another. See the `--config` option.

To help booting the first node a DHCP server is started which is configured to listen **only** on the test bridge. Once the first node is running this server can then be shut down to allow the first node to take over the DHCP function (`--shutdowndhcp`).

**NOTE:** [NAT](https://en.wikipedia.org/wiki/Network_address_translation) is not currently supported. NAT is needed if the nodes need to communicate outside the virtual test network. This may be supported in a future version.

start-vm
--------

The `start-vm` script uses [KVM](http://wiki.qemu.org/KVM) to run virtual machines. A big reason for KVM is it supports nested virtual machines provided your [kernel supports it](https://fedoraproject.org/wiki/How_to_enable_nested_virtualization_in_KVM). Installing QEMU-KVM is outside the scope of this document. Look for instructions relevant to the distro on which you will be running VMs.

**NOTE:** Each VM requires approximately 3GB of RAM.

After using `vm-network` to configure the network for a test scenario the VMs can be started using the `start-vm` script. One VM per *tapset* can be started. Each VM is assigned its own [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) with the last byte being the same as the *tapset* number used for the VM.

Even though a given *tapset* may contain a large number of TAP interfaces a VM need only use a subset of those interfaces. This is controlled using the `--numvmif` option. Each of the interfaces used by the VM is given a unique MAC address again derived using the *tapset* number as part of the MAC address and using a scheme similar to the TAP interfaces but with the 5th byte having the pattern `8<n>` where `<n>` is the *tapset* number. This avoids conflicts with the MAC addresses assigned to the TAP interfaces while at the same time providing information making it easy to identify corresponding interfaces. **NOTE:** This scheme effectively limits the practical maximum number of VMs to 9 (1 thru 9).

Each VM can be started using images from a local build or downloaded from a build server which defaults to [S3](http://omniti-cerana-artifacts.s3.amazonaws.com/index.html?prefix=CeranaOS/jobs/build-cerana/).

Booting either kernel and initrd images or using an ISO is possible.

This script creates one or more disk images (`--numdisks`) for each VM which helps verification of [ZFS](https://en.wikipedia.org/wiki/ZFS) within Cerana. Each disk image by default is given a name which also uses the *tapset* to help identify it. This naming scheme by default uses the pattern `sas-<tapset>.<n>` where `<n>` is the disk number. For example a configuration having three VMs (*tapsets*) with two disk images each produces images having the following names:

```
        sas-1-1.img
        sas-1-2.img
        sas-2-1.img
        sas-2-2.img
        sas-3-1.img
        sas-3.2.img
```

Examples
========

Single Node
-----------

This example shows what happens when using only the default values. First `cd` to the directory where you want the various images (i.e. kernel, initrd, disk) to be saved.

```
vm-network --verbose
```

**NOTE:** If you've already been running `vm-network` you may want to use the `--resetdefaults` option to return to a known default state.

The interfaces `ceranatap.1.1` and `ceranabr.1` were created and `ceranatap.1.1` added to the `ceranabr.1` bridge. The `ceranabr.1` bridge was assigned the IP address `10.0.2.2`. The `dhcpd` daemon was started and configured to listen only on the `10.0.2.0` subnet.

A configuration named `single` was created and saved in the `~/.testcerana` directory.

In this case the `artifacts` directory must exist and contain the kernel and initrd images.

```
start-vm --verbose
```

**NOTE:** If you've already been running `start-vm` you may want to use the `--resetdefaults` option to return to a known default state.

The interface `ceranatap.1.1` is used as the management interface and by virtue of the `dhcpd` daemon is assigned the IP address `10.0.2.200`.

A disk image named `sas-1-1.img` was created and uses as a virtual disk for the VM.

The boot messages will have spewed to the console and you can log in as the root user (no password).

The root prompt contains the UUID assigned to the VM with the last byte equal to "01" to indicate the single VM in this scenario.

**NOTE:** Use [`^AX`](http://qemu.weilnetz.de/qemu-doc.html#mux_005fkeys) keystrokes to shutdown the VM.

Two Nodes
---------

This example requires reconfiguring the network to support two *tapsets* because of using two nodes in the test scenario. It also shows booting an ISO instead of the normal kernel images.

```
vm-network --verbose --numsets 2 --config two-node
```

Because this example is also booting the Cerana ISO it is necessary to shutdown the `dhcpd` daemon so that Cerana can take over that function.

```
vm-network --verbose --shutdowndchpd
```

This saves another configuration named `two-node` in the `~/.testcerana` directory. The existing network configuration was torn down and the new one created. The interfaces `ceranatap.1.1` and `ceranatap.2.1` have been created and linked to the `ceranabr.1` bridge.

```
start-vm --verbose --boot iso
```

This boots the ISO image which in turn displays the GRUB menu. Cursor down and select the "CeranaOS Cluster Bootstrap" option. This starts the first node of the two node cluster.

Now open another console and `cd` to the same directory. This time use `start-vm` to start a second VM.

```
start-vm --verbose --boot iso --tapset 2
```

This too boots the ISO images which again displays the GRUB menu. This time cursor down and select the "CeranaOS Cluster Join" option.

This causes the second node to use the PXE protocol to boot using images provided by the `bootserver` running on the first node. Its management interface was assigned an IP address by the `dhcp-provider` also running on the first node.

Adding Interfaces to an Existing Configuration
----------------------------------------------

There are times when an additional interface is needed either to support an additional node or to add an interface for a node. This is a relatively simple thing to do using `vm-network`. Using the "Single Node" example above adding an interface is as simple as:

```
vm-network --verbose --numsets 2 --config single
```

This creates another TAP interface, `ceranatap.2.1`, for the `single` configuration and adds it to the `ceranabr.1` bridge. It was not necessary to shutdown the test network in this case. This also illustrates that `vm-network` is able to repair a network configuration (within limits) if an interface was deleted for any reason.

**NOTE:** At this time removing interfaces is possible but the script will not automatically delete TAP interfaces if the new number is smaller than before (e.g. the new `--numsets` is 1 but was 2). This a feature for the future. However, this does not cause a problem because the script will tear down all interfaces linked to a bridge when switching configurations or when the `--shutdown` option is used.

Using Downloaded Images
-----------------------

The `start-vm` script also supports downloading and using specific builds from a server. By default it downloads from the [CeranaOS instance on Amazon S3](http://omniti-cerana-artifacts.s3.amazonaws.com/index.html?prefix=CeranaOS/jobs/). This requires use of the `--job` and the `--build` options. The `--job` option defaults to `build-cerana` and is good for most cases. The `--build` option however has no default and must be set to a valid number ([look on S3](http://omniti-cerana-artifacts.s3.amazonaws.com/index.html?prefix=CeranaOS/jobs/build-cerana/)) before the download will work. For example the following downloads and boots build 121. All other options are whatever were set in a previous run.

**NOTE:** Symlinks are used to point to the build to boot. If files exist this script will not removed them. If you want to use the same directory you will need to manually remove the images.

```
start-vm --verbose --download --build 121
```

A Three Node Demo
-----------------

This example creates a three node cluster for demonstrating communication of Cerana components between the nodes. Running the demo is described in the [demo documentation](https://github.com/cerana/cerana/blob/demo/docs/demo/README.md). Refer to that document but use these steps to [bootstrap the cluster](https://github.com/cerana/cerana/blob/demo/docs/demo/README.md#bootstrap-the-cluster).

This example is for running the demo using KVM in a Linux environment so the Linux version of the [coordinator-cli](http://omniti-cerana-artifacts.s3.amazonaws.com/CeranaOS/testdata/coordinator-cli) is needed.

It's recommended that the demo be started in an empty directory. The [demo documentation](https://github.com/cerana/cerana/blob/demo/docs/demo/README.md) lists some files which are required files run the demo. You can download the files using the links provided here. You should have these files:

-	[backend-in-a-zfs-stream.zfs](http://omniti-cerana-artifacts.s3.amazonaws.com/CeranaOS/testdata/backend-in-a-zfs-stream.zfs)
-	[haproxy-in-a-zfs-stream.zfs](http://omniti-cerana-artifacts.s3.amazonaws.com/CeranaOS/testdata/haproxy-in-a-zfs-stream.zfs)
-	[coordinator-cli](http://omniti-cerana-artifacts.s3.amazonaws.com/CeranaOS/testdata/coordinator-cli) (**NOTE:** This is the Linux version)
-	[setup-application.sh](https://raw.githubusercontent.com/cerana/cerana/demo_outline/docs/demo/setup-application.sh)

Running the demo requires a bridge with an IP address on the 172.16 subnet and no DHCP server. If you haven't done so first shutdown the existing network configuration.

```
vm-network --shutdown
```

Now the network can reconfigured for the demo. Because of running three nodes in the demo three *tapsets* are required. The bridge IP address also needs to be configured for the demo. For this demo CeranaOS provides its own DHCP server don't start the DHCP server.

```
vm-network --numsets 3 --bridgeip 172.16.2.2 --maskbits 16 --nodhcpd  --config three-node-demo
```

Four separate consoles are needed to run the demo. One for each node and one to interact with the cluster. For this example these are referred to as `node1`, `node2`, `node3` and `shell`.

### node1

The first node is booted using the `cerana.iso` image which can be downloaded from [S3](http://omniti-cerana-artifacts.s3.amazonaws.com/index.html?prefix=CeranaOS/jobs/build-cerana/). It's recommended the latest build be used for this example.

```
start-vm --download --build 132 --boot iso --tapset 1
```

When the GRUB menu appears select the "CeranaOS Cluster Bootstrap" option. This will boot using the kernel and initrd images from `cerana.iso`.

You can verify the network configuration using the node console. Login as root (no password) and:

```
[root@144cae4e-ff5c-dd43-ad51-076385611e01:~]# ip address
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: mgmt0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether de:ad:be:ef:81:01 brd ff:ff:ff:ff:ff:ff
    inet 172.16.10.2/16 brd 172.16.255.255 scope global mgmt0
       valid_lft forever preferred_lft forever
    inet6 fe80::dcad:beff:feef:8101/64 scope link
       valid_lft forever preferred_lft forever
```

### node2

The second node uses the `dhcp-provider` and `bootserver` from `node1` to boot and requires using the second *tapset*. This boots using images served from `node1`.

```
start-vm --boot net --tapset 2
```

When the login prompt appears login as root and:

```
144cae4e-ff5c-dd43-ad51-076385611e02 login: root

[root@144cae4e-ff5c-dd43-ad51-076385611e02:~]# ip address
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: mgmt0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether de:ad:be:ef:82:01 brd ff:ff:ff:ff:ff:ff
    inet 172.16.219.100/16 brd 172.16.255.255 scope global dynamic mgmt0
       valid_lft 86345sec preferred_lft 86345sec
    inet6 fe80::dcad:beff:feef:8201/64 scope link
       valid_lft forever preferred_lft forever
```

Note the MAC and IP addresses for this node. The `82` in the MAC address indicates this is the second node and the IP address was provided by the `dhcp-provider` running on `node1`. Also, the last two characters in the hostname (UUID in the prompt) are `02` also indicating this is the second node.\`

### node3

Starting the third node is similar to the second but uses the third *tapset*.

```
start-vm --boot net --tapset 3
```

Again:

```
144cae4e-ff5c-dd43-ad51-076385611e03 login: root

[root@144cae4e-ff5c-dd43-ad51-076385611e03:~]# ip address
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: mgmt0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether de:ad:be:ef:83:01 brd ff:ff:ff:ff:ff:ff
    inet 172.16.63.98/16 brd 172.16.255.255 scope global dynamic mgmt0
       valid_lft 86386sec preferred_lft 86386sec
    inet6 fe80::dcad:beff:feef:8301/64 scope link
       valid_lft forever preferred_lft forever

```

Note the MAC and IP addresses for this node. The `83` in the MAC address indicates this is the third node and the IP address was provided by the `dhcp-provider` running on `node1`. Also, the last two characters in the hostname (UUID in the prompt) are `03` also indicating this is the third node.

### shell

In the `shell` console verify the nodes are properly connected to the bridge.

Ping the first node using the address disovered for `node`. For this example `node1` has the address `172.16.10.2`. This, by the way, was statically assigned by the GRUB menu.

```
$ ping 172.16.10.2
PING 172.16.10.2 (172.16.10.2) 56(84) bytes of data.
64 bytes from 172.16.10.2: icmp_seq=1 ttl=64 time=1.31 ms
64 bytes from 172.16.10.2: icmp_seq=2 ttl=64 time=2.02 ms
64 bytes from 172.16.10.2: icmp_seq=3 ttl=64 time=0.442 ms
```

Now do the same for the second node. In this example the address was assigned by the `dhcp-provider` running on `node1` and is `172.16.219.100` (from `node2` above).

```
$ ping 172.16.219.100
PING 172.16.219.100 (172.16.219.100) 56(84) bytes of data.
64 bytes from 172.16.219.100: icmp_seq=1 ttl=64 time=1.10 ms
64 bytes from 172.16.219.100: icmp_seq=2 ttl=64 time=0.379 ms
64 bytes from 172.16.219.100: icmp_seq=3 ttl=64 time=0.414 ms
```

And finally, the third node.

```
$ ping 172.16.63.98
PING 172.16.63.98 (172.16.63.98) 56(84) bytes of data.
64 bytes from 172.16.63.98: icmp_seq=1 ttl=64 time=0.572 ms
64 bytes from 172.16.63.98: icmp_seq=2 ttl=64 time=0.384 ms
64 bytes from 172.16.63.98: icmp_seq=3 ttl=64 time=0.348 ms
```

### Run The demo

All three nodes are now ready to run the demo described in the [demo documentation](https://github.com/cerana/cerana/blob/demo/docs/demo/README.md).

More
----

From time to time more examples will be added to this document to illustrate progressively complex scenarios.
