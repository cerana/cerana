Planning diagrams and documentation
=================

Boot Process Diagram
--------------------

![Boot Process Diagram](https://rawgithub.com/cerana/cerana/master/docs/planning/boot.svg)

Kernel Command Line
-------------------
* “cerana.mgmt_mac” MAC address to be used for IP address  - the one used for PXE
* “cerana.mgmt_ip” IP address in CIDR format (IP/netmask)
* “cerana.mgmt_gateway” ? Gateway ? (optional)
* “cerana.zfs_config={manual,auto,auto-mirror,auto-raidz}” ZFS Pool Info (auto or not) default to manual
* “cerana.cluster_ips=host1,host2,host3” Cluster info -  List of cluster IPs (host1,host2,host3), check for cluster info in pool, if nothing, then no cluster
* “cerana.cluster_bootstrap” - When present this machine should become the bootstrap node of a cluster (install all layer 2 services and act as DHCP/PXE server for subsequent nodes)
* “cerana.rescue” Rescue mode - do not import ZFS pool.

File System Layout
------------------

* tmpfs /task-socket/node-coordinator/
* tmpfs /task-socket/cluster-coordinator/

* zpool $zpool=”data”
* /$zpool/datasets/ro/${dataset-id}/stuff
* /$zpool/datasets/rw/${dataset-id}/stuff
* /$zpool/running-clones/${clone-id}
* /$zpool/config
* /$zpool/services
* /$zpool/logs
* /$zpool/platform/${version}/{bzImage,initrd}
* /$zpool/platform/grub.conf

* /etc/systemd-mutable/system (writable/mutable directory NixOS systemd will look for unit files. we must create it ourselves)

Software in image:
------------------
* Bash, coreutils
* Grub2
* Systemd
* Daisy binary
* Go binaries
* Qemu-kvm
* Init-scripts
* Ntp
* Whichever binaries the hardware identification task provider needs
* Initial Debugging tools: strace, gdb
* ebpf tools?

Ideally temporary:
* SSH
