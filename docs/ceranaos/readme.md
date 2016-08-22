CeranaOS is a lightweight in-memory distribution of linux designed to securely and easily run isolated container and kvm processes. The design allows for management and information retrevial of the node to occur without logging in, through the use of a node-local Cerana Task Api and a secure process running platform Daisy.

CeranaOS is a standalone operating system, but was desgined with the goals of Cerana Platform in mind.

CeranaOS uses ZFS as the storage platform. ZFS allows for quick clones of images and datasets, which is a perfect use case for both stateless services (fresh cost-nothing clone of the backing image at time of running), and stateful services (mounting in a permanent datastore into your running container, cloned or not). 

All CeranaOS platform software and user containers/kvm process run as services. If it is declared on the box, it is run.
