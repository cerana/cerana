# OS-level virtualisation, background

---

## Prehistory

---

### Multiple OS instances interacting with hypervisor
* CP-40 through Xen

### chroot
* Traditional process isolation through restricted filesystem view, other resources shared

---

## History

---

## FreeBSD jails (Kamp & Watson, 2000)

### Motivation:
* Delegate administrative functions to untrusted parties
* Impose system-wide mandatory policies on process interaction and sharing
* Avoid userspace impact and administrative complexity of fine-grained approaches

### Implementation:
* "Partitioning" - filesystem, network and process resources are assigned to a management environment
* Entire base system image, or any subset thereof in each filesystem environment
* Single system call to create jails, jail context attached to processes via inheritance
* Processes within a jail see a restricted view of fs/net/proc, even with root privileges
* Limited set of syscalls allowed to a jailed processes 
* "Jail awareness" also applies to some device drivers
* Later additions: Per-jail resource limits: memory, CPU
* Later additions: Full network stack isolation with vnet

### Limitations:
* Needs dedicated IP address
* Only subset of devices can be safely exposed

---

## Linux-VServer (Ligneris, 2005)

### Motivation:
* Better resource utilisation through hardware consolidation
* Support "grid computing" deployments
* Support dynamic allocation of resources

### Design:
* New syscalls mirroring chroot, contexts attached to processes via inheritance
* chcontext for process isolation (with contexts 0 and 1 having special meaning)
* chbind for net isolation (bound to a specific IP address)
* POSIX capabilities to limit resource access to privileged processes

### Limitations:
* All privileges have same meaning, but processes see a subset of select namespaces
* Relies on POSIX capabilities only for all other syscall restrictions
* Needs dedicated IP address
* No specific device driver support

---

## Solaris zones (Price & Tucker, 2004)

### Motivation:
* Consolidation of existing server workloads which may require dedicated named resources
* Isolation from privilege escalation and resource exhaustion in compromised environments
* Low resource overhead
* Simple adminstrative interface

### Design:
* Zone per workload, encompassing filesystem, network, process, IPC and other resources
* All process privileges are interpreted in zone context
* Persistent configuration for zones
* Fine-grained resource management available
* New APIs and configuration allows delegation of resources to non-global zone
* Later additions: Branded zones supporting alternative kernel personalities
* Later additions: Whole IP stack zones and Crossbow network virtualisation
* Later additions: I/O throttling

### Limitations:
* Needs dedicated IP address
* Only subset of devices can be safely exposed

---

## OpenVZ

### Motivation:
* Zones-like support for isolated environments in Linux

### Design:
* Filesystem, IPC, user ID, process, network resource isolation
* Two-level quotas, I/O and CPU schedulers
* Container checkpoint and migration

### Limitations:
* No specific device driver support

---

# Linux isolation and resource management pieces

---

## Linux namespaces

### Motivation:
* Need set of primitives for sandboxing specific resources

### Design:
* Namespaces are named resources
* Processes can be created in a specified namespace, or moved after creation
* IPC namespace isolates distinct set of IPC resources
* Network namespace isolates IP stack, network devices
* Mount namespace isolates available mountpoints - extended chroot
* PID namespaces with special process 1 semantics, can be nested
* User namespaces (3.12+) translate UID/GID ranges inside/outside the namespace

### Limitations:
* Useful as primitives for new applications, no resource limits
* Complex to get correct set of privileges in processes under child namespaces
* Management stack not included
* User namespace file access checks are mapped to initial namespace; entire mapping range must be assigned to an "owner" process in parent
* No specific device driver support

---

## Linux control groups (cgroups)

### Motivation:
* Need for resource control and accounting of process trees

### Design:
* Provides interface to specify heirarchical set of resouce constraints and monitor usage
* Can be integrated with namespace isolation for managing processes within namespaces
* Support for process checkpoint and migration

### Limitations:
* To avoid race conditions, must have single privileged process manage all constraints

---

## Checkpoint/Restore In Userspace (CRIU)

* Spinoff of OpenVZ for checkpointing and migration of process state, using minimal new syscalls

---

Linux Containers (LXC)
----------------------
* Integrates the three above technologies with configuration store
* Provides administrative tools for container management

libcontainer
------------
* Lightweight container management, using namespaces and cgroups
* Provides Go API for creating and managing containers

---

## Towards automation

---

## SmartOS/SDC

* API and image-based tools for zone creation and management
* KVM branded zones run single process with externally supplied images

## LXD

* Image-based API for container management and migration
* Wrapper for LXC

## Docker

* High-level API for deploying applications inside containers
* Filesystems constructed on the host, on the fly
* Automation for creating distributed systems with containers
* Backends for LXC, libvirt, and libcontainer (created by Docker)

---

# Network virtualisation, background

---

## Crossbow (Tripathi et. al., 2009)

### Motivation:
* Create virtual network interfaces with hardware acceleration
* Provide scheduling of network I/O
* Allow separate isolated admistrative domains

### Implementation:
* "Virtualization lanes" - pools of dedicated network hardware and CPU resources bound to virtual network interfaces (VNICs)
* "Etherstubs" can be used instead of hardware NICs
* Virtual switches are implicitly created to support VNICs, can be distributed across hosts
* Switching engine provides I/O prioritisation and scheduling
* VNICs can be delegated to zones for management; address spoofing is prevented

---

## Open vSwitch (Pfaff et. al., 2009)

### Motivation:
* Provide software-based switching between virtual environments

### Implementation:
* Multi-layer switch supports common features and protocols of hardware switches
* Kernel and userspace forwarding engines
* High level configuration for virtual network distributed across multiple switches
* Integrated into Linux, supports FreeBSD and NetBSD
