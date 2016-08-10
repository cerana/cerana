![Cerana](https://rawgithub.com/cerana/cerana/master/docs/logos/cerana_logo_side.svg)

# Demo time!

## Requirements

* Dedicated network (no existing DHCP server)
* 3 machines / VMs
* Control machine with coordinator-cli, jq, curl, tr, and tee that can talk to the cluster network (in the 172.16.0.0/16 subnet)

If you are on SmartOS for your control machine, you can use these prebuilt [coordinator-cli](https://us-east.manta.joyent.com/nahamu/public/cerana/demo/coordinator-cli-smartos) and [jq](https://us-east.manta.joyent.com/nahamu/public/cerana/demo/jq-smartos) binaries.

If you are on Linux, you can obtain coordinator-cli from a CeranaOS machine, or use the copy available here: [coordinator-cli](https://us-east.manta.joyent.com/nahamu/public/cerana/demo/coordinator-cli-linux)

## Bootstrap the cluster

1. Boot first machine from CD/ISO, select the Bootstrap option
1. Wait for it to come up and present a login prompt
1. Log in and verify / wait until /data/platform/current symlink exists
1. Boot second and third machines over network
1. Wait long enough for them to finish downloading platform images and creating their own /data/platform/current symlinks

## Run the demonstration application stack

1. Get images for the components of the demo application stack and the script to talk to the coordinator
  * [Backend ZFS Stream](https://us-east.manta.joyent.com/nahamu/public/cerana/demo/backend-in-a-zfs-stream.zfs)
  * [Haproxy ZFS Stream](https://us-east.manta.joyent.com/nahamu/public/cerana/demo/haproxy-in-a-zfs-stream.zfs)
  * [Demo Script](https://raw.githubusercontent.com/cerana/cerana/demo_outline/docs/demo/setup-application.sh)
1. Run script to upload images into cluster and create services and bundles
```bash
alias l2-request='coordinator-cli -c http://172.16.10.2:8085 -r :4080'
# repeat the following until you see all three nodes:
l2-request -t kv-keys -a key=/nodes | jq
./setup-application.sh
```
1. Run the demo-tick (on one of the nodes)
```bash
run-demo-tick
```
1. Verify that the application is working
```
l2-request -t kv-keys -a key=/nodes | tr ',' '\n' | tr -d '][nodes/"' | while read ip; do for i in {1..3}; do curl http://$ip:8001/$ip; done; done;
```

## ZFS Delegation Bonus Demo (waiting for Daisy)

1. Use daisy to create/enter a container
1. Delegate a ZFS filesystem to it
1. Enter container
1. Do ZFS operations

## Cluster Recovery Demo
1. Shut down first machine cleanly
1. Roll it back to pristine / destroy and recreate it.
1. Boot first/fourth machine from network
