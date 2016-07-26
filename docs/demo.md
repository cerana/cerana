![Cerana](https://rawgithub.com/cerana/cerana/master/docs/logos/cerana_logo_side.svg)

# Demo time!

## Requirements

* Dedicated network (no existing DHCP server)
* 3 machines / VMs

## Bootstrap the cluster

1. Boot first machine from CD/ISO, select the Bootstrap option
1. Wait for it to come up and present a login prompt
1. Log in and verify / wait until /data/platform/current symlink exists
1. Boot second and third machines over network
1. Wait long enough for them to finish downloading platform images and creating their own /data/platform/current symlinks
1. Shut down first machine cleanly
1. Roll it back to pristine / destroy and recreate it.
1. Boot first/fourth machine from network

## Run the demonstration application stack

1. Get images for the components of the stack
1. Upload images into cluster
1. Define the service bundles
1. Run the demo-tick

Cool commands to know

* Coordinator CLI commands

```bash
#/bin/bash

set -o xtrace

kv-decode () { json data | base64 -d | json; }
l2-request () { coordinator-cli -c http://172.16.10.2:8085 -r 172.16.250.250:4080 "$@"; }
l1-request () { coordinator-cli -c http://172.16.10.2:8080 -r 172.16.250.250:4080 "$@"; }

service () {
cat <<EOF
{
  "service": {
    "dataset": "${1}",
    "cmd": [
      ${2}
    ],
    "env": {
      "PORT": "${3}"
    }
  }
}
EOF
}

bundle () {
cat <<EOF
{
  "bundle": {
    "services": {
      "${1}": { "id": "${1}" }
    }
  }
}
EOF
}

DATASET=$(l2-request -t import-dataset -s -a redundancy=1 < backend-in-a-zfs-stream.zfs | tee /dev/stderr | json dataset.id)
SERVICE=$(service ${DATASET} '"/backend"' 10000 | l2-request -t update-service -j | tee /dev/stderr | json service.id)
bundle ${SERVICE} | l2-request -t update-bundle -j

DATASET=$(l2-request -t import-dataset -s -a redundancy=1 < haproxy-in-a-zfs-stream.zfs | tee /dev/stderr | json dataset.id)
SERVICE=$(service ${DATASET} '"/haproxy.py"' 8001 | l2-request -t update-service -j | tee /dev/stderr | json service.id)
bundle ${SERVICE} | l2-request -t update-bundle -j
```

```
cat backend-in-a-zfs-stream.zfs | coordinator-cli -s -t zfs-receive -c http://172.16.10.2:8080 -r 172.16.250.250:4080 -a name=data/datasets/backend-test1
cat backend-in-a-zfs-stream.zfs | coordinator-cli -s -t import-dataset -c http://172.16.10.2:8085 -r 172.16.250.250:4080 -a redundancy=3
coordinator-cli -c http://172.16.10.2:8085 -r 172.16.250.250:4080 -t kv-keys -a key=/
```

* Commands on a node
```
journalctl -o json -f -n 100 | jq -C '{service: .SYSLOG_IDENTIFIER, log: .MESSAGE|fromjson }' 2>/dev/null | less -R
demotick -c unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock -d data/datasets -n 8080 -r 10s -l debug 2>&1 | jq . -C | less -R

```

## ZFS Delegation Bonus Demo

1. Use daisy to create/enter a container
1. Delegate a ZFS filesystem to it
1. Enter container
1. Do ZFS operations
