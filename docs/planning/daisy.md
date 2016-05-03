# "Daisy" Binary:

The "daisy" binary will be invoked by a systemd unit to launch processes in the appropriate namespaces and with the appropriate filesystems.
It will be used both for launching "user" processes, but also to launch our task providers.

* open connection to coordinator
* open a response socket
* ask coordinator what network namespace should I use?
* set the network namespace
* unshare UTS, IPC, delegation, mount, user, PID namespaces (ZFS provider won't unshare delegation)
* pivot_root to the new root filesystem (stay at / for taks providers)
* set selinux policy
* ask coordinator for a unique mapping for the root user - some other process then does that to this process
* set up cgroups
* set seccomp policy
* close the socket descriptors (outbound to coordinator and inbound response sockeet)
* drop most capabilities
* "reset root" - set the uid of my pid to 0 (which actually gets mapped)
* drop remaining capabilities that need dropping
* execve the intended process
