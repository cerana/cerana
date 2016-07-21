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

## ZFS Delegation Bonus Demo

1. Use daisy to create/enter a container
1. Delegate a ZFS filesystem to it
1. Enter container
1. Do ZFS operations
