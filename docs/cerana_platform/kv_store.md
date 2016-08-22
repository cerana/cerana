Data Architecture Tree (WIP)
* dataset 
    * id: uuid 
    * parent: uuid to be replicated from 
    * parent_same_machine: boolean 
    * read-only: boolean 
    * nfs: boolean 
    * redundancy: integer 
    * quota: integer 
* service 
    * id: uuid 
    * dataset: uuid 
    * cmd: array of strings 
    * /healthcheck 
        * /{id} 
            * {protocol_provider}: parameters 
    * /limits 
        * Cpu: 
        * Memory: 
        * Processes: 
        * ... 
    * /env 
        * ENVVAR_NAME: value 
* bundles/ 
    * id: integer 
    * /datasets 
        * /{your-name} 
            * uuid: Optional dataset-uuid 
            * type: rw zfs, temp zfs or ramdisk 
            * quota: optional quota 
    * /services 
        * /{service-uuid} 
            * (override anything but id and dataset) 
            * /limits 
                * [optional overrides] 
            * /env 
                * ENVVAR_NAME: value 
            * /datasets 
                * {your-name}: mount_point,rw 
    * redundancy: integer 
    * ports 
        * /{port-number} 
            * Connected_bundles: array of ids 
* external-port 
    * /{port-number} 
        * Bundle: bundle_id 
        * Port: port_from_bundle 
* /nodes 
    * /{machine-serial} TTL’d (ip address for this, or maybe mac) (self assigned ip config) 
        * Heartbeat: last write 
        * Memory_free 
        * Cpu_free 
        * Disk_free 
        * Memory_total 
        * Cpu_total  
        * Disk_total 
* /historical_entries 
    * /{machine-serial} 
        * /{heartbeat} 
            * The stuff under nodes 
* /cluster_config 
    * zfs: auto/manual 
* /heartbeat 
    * /bundles 
        * /id (ephemeral directory) 
            * “serial”: “ip” (TTL Record) 
    * /datasets 
        * /id (ephemeral directory) 
            * “ip”: boolean indicating if in use (TTL Record)
