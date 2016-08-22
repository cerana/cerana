Cerana Platform

Cerana Platform is an easy to stand up and use clustering platform for a set of CeranaOS nodes. Cerana Platform is built off a few core principals:
- Declarative
- Eventual Consistency
- It is okay to fail. All things should be coded to expect failure in actions

Installation Instructions and demo running (WIP):
- https://github.com/cerana/cerana/tree/demo_outline/docs/demo
- https://github.com/cerana/cerana/tree/321_kvm_demo_fixes/build/scripts

Cerana Platform is comprised of the following components:
- Key Value Store (Consul Today)
- Cluster Level Cerana Task API
- Cluster Ticks
- Node Ticks
- Bootserver

Key Value Store

The KV store is the heart of the Declarative engine. Things that are declared are stored here, and the rest of the cluster software tries to make this a reality. The KV Store has configuration for other pieces of clusters, plus a few top level objects of note:
- Datasets: A top level object representing a read only or read write dataset. Read only datasets currently have redundancy across the cluster.
- Services: A process executed using a dataset as the backing store. Can be a container or kvm vm (kvm is todo)
- Bundles: Bundles are groups of confuration optionally with services. As the system load balances layer 3, bundles can be useful for pure networking.

Cluster Level Cerana Task API

This is a Cerana Task API that runs a coordinator and providers on every node in the cluster, giving it ultimate redundancy and short pathing from other pieces of software. All changes and information retrievels go through tasks, and all other software manipuatling the Cerana Platform make changes through way of the coordinator, instead of directly. Though this, we can define top level objects in KV store, as well as import datasets. Also internally, this is used to maintain dhcp for not only the cluster, but all the service networks.

List of providers in in cluster-coodinator
- https://github.com/cerana/cerana/tree/master/providers/datatrade
- https://github.com/cerana/cerana/tree/master/providers/dhcp
- https://github.com/cerana/cerana/tree/master/providers/kv

Cluster Ticks

These are responsible for assuring the desired state becomes the reality. These are installed thoughout the cluster, but run on a single cluster only. Using the Cluster Coodinator, this reads the state of the KV, then talks to various node coordinators to add and remove services, bundles, and datasets. (WIP, see demo tick in demo)

Node Ticks

These are ticks that run on every node that push stats about that node, as well as manipulate that node based on definitions (such as assuring the network stack for services). Currently this is in the form of statspusher. Splitting apart to separate ticks is a WIP.
- https://github.com/cerana/cerana/tree/master/cmd/statspusher

Bootserver

Bootserver provides cluster-redundant dhcp and related services (tftp/http).
- https://github.com/cerana/cerana/tree/master/cmd/bootserver

Service-Oriented Networking

(Docs coming soon)
