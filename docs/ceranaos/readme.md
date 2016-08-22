CeranaOS

CeranaOS is a lightweight in-memory distribution of linux designed to securely and easily run isolated container and kvm processes. The design allows for management and information retrevial of the node to occur without logging in, through the use of a node-local Cerana Task Api and a secure process running platform Daisy.

CeranaOS is a standalone operating system, but was desgined with the goals of Cerana Platform in mind.

CeranaOS uses ZFS as the storage platform. ZFS allows for quick clones of images and datasets, which is a perfect use case for both stateless services (fresh cost-nothing clone of the backing image at time of running), and stateful services (mounting in a permanent datastore into your running container, cloned or not). ZFS also allows secure ZFS delegation for advanced storage setup within a container or kvm process.

All CeranaOS platform software and user containers/kvm process run as services. If it is declared on the box, it is run. Processes are executed using the Daisy binary. This binary allows us to properly set up namespaces and zfs, as well as set secomp and selinux policies (including default policies for easy secure running)
https://github.com/cerana/cerana/tree/daisy/cmd/daisy

Manipulation of CeranaOS should occur through Cerana Task API. CeranaOS comes with both easy abstracted tasks that quickly allow you to run your services, as well as underlying platform tasks beneath (ex, systemd and zfs) for more advanced manipulation.
List of current providers for the node-coodinator
https://github.com/cerana/cerana/tree/master/providers/health
https://github.com/cerana/cerana/tree/master/providers/metrics
https://github.com/cerana/cerana/tree/master/providers/namespace
https://github.com/cerana/cerana/tree/master/providers/service
https://github.com/cerana/cerana/tree/master/providers/systemd
https://github.com/cerana/cerana/tree/master/providers/zfs

Interesting Reads:
https://github.com/cerana/cerana/tree/master/docs/planning
