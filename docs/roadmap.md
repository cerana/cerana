Roadmap and motivations beginning 8/15/2016

After a long road, we have brought all the pieces together and have created a running demo of CeranaOS and Cerana Platform. Executing containers across the platform works, and the demo tick runs with great gusto.

What has been completed:
- A fully async Cerana Task API, a resilient and recursive task stack that works statelessly
- CeranaOS as a lightweight, in-memory linux os for easily running containers and kvm processes
- Cerana Platform, a declarative engine for assuring the state of a cluster of CeranaOS Boxes. As of today it has a resilient key-value store, a demo "tick" to assure the state of services and images across the cluster, and a Cerana Task API for executing tasks to change the state of the cluster.

What comes immediately next (two sprints):
- Creating initial Cerana Platform ticks that run on a schedule, replacing the demo tick
 - Service Bundle tick
 - Dataset Assurance Tick
- Superseding Stats-pusher with the notion of per node ticks
 - Bundle Heartbeat Pusher
 - Dataset Heartbeat Pusher
 - Node Heartbeat Pusher
 - Create Network Namespace assurance tick
 - Create Service DHCP tick
- Cerana Platform Networking.
 - The service oriented model and L3 as a service planning (done)
 - Expanding our dhcp provider to allow for dhcp buckets (for allowing different service_bundle networks to have unique addressing across the cluster)
 - Creating a networking provider in CeranaOS for standing up networking namespaces as needed, while abstracting our philosophies
 - A Networking Tick in Cerana Platform on each box for handling needed network namespace standing, dhcp-like leasing using the dhcp-provider, and firewalling based on service allowance
 - A Netfilter provider in layer 2 for generating iptables policy based on service_bundle permission access

What is next (rest of Q4 goals):
- Accept compressed archive as dataset imports (today we accept zfs)
- Advanced logic for service bundle placement
- Log retrieval
- Clean up and enable health-checks
- Create optimal dataset garbage collection
- Add ZFS feature for mapping arbitrary user operations to root
- Website and clean documentation
- Shrinking CeranaOS image
- Demo software using Cerana Platform
 - Simple user access based cloud provider
 - Docker endpoint for cluster
 - Data-lake software test
