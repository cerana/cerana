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

Testing interaction between the various nodes of a Cerana cluster requires a network configuration which allows communication between the nodes but avoids flooding the local network with test messages. This is accomplished using a bridge to connect a number of [TAP](https://en.wikipedia.org/wiki/TUN/TAP) devices. When using this script it helps to keep the following in mind. * A Cerana cluster can be comprised of 1 or more VMs. Each VM can have 1 or more TAP interfaces. The TAP interfaces to be associated with a specific VM is termed a "tapset". The number of *tapsets* is controlled using the option `--numsets`.

By default TAP interfaces are named using using the pattern `tap.<tapset>.<n>` where `<n>` is TAP number within a *tapset*. For example a configuration having three VMs with two interfaces each produces three *tapsets* each having two interfaces. The resulting TAP interfaces become:

```
        tap.1.1
        tap.1.2
        tap.2.1
        tap.2.2
        tap.3.1
        tap.3.2
```

Each TAP interface is assigned a MAC address beginning with the default pattern "DE:AD:BE:EF". The 5th byte of the MAC address is the number of the corresponding *tapset* and the 6th byte is the TAP number within the *tapset*.

These are then all linked to a single bridge having the default name `ceranabr0`.

**NOTE:** Currently only a single bridge is created. In the future multiple bridges will be used to better support testing VMs having multiple TAP interfaces. For example one bridge can be used for the node management interfaces while a second bridge can be used for a connection to a wider network.

The `vm-network` script also supports maintaining multiple network configurations making it easy to tear down one configuration and then setup another. See the `--config` option.

To help booting the first node a DHCP server is started which is configured to listen **only** on the test bridge. Once the first node is running this server can then be shut down to allow the first node to take over the DHCP function (`--shutdowndhcp`).

**NOTE:** [NAT](https://en.wikipedia.org/wiki/Network_address_translation) is not currently supported. NAT is needed if the nodes need to communicate outside the virtual test network. This may be supported in a future version.

start-vm
--------

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