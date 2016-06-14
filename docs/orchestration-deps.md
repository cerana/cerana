## Node Layer

- coordinator
- health-provider
    - systemd-provider
- metrics-provider
- statspusher
    - clusterconf-provider
    - metrics-provider
    - systemd-provider
    - zfs-provider
- systemd-provider
    - systemd
- zfs-provider
    - zfs (ClusterHQ stable ioctl API version)

## Cluster Layer

- clusterconf-provider
    - kv-provider
- kv-provider
    - consul