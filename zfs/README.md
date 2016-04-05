# Pure Go Client for ZFS Stable API #

[![GoDoc](https://godoc.org/github.com/mistifyio/gozfs?status.svg)](https://godoc.org/github.com/mistifyio/gozfs)
![status=alpha](https://img.shields.io/badge/status-alpha-orange.svg)
[![MIT Licensed](https://img.shields.io/github/license/mistifyio/gozfs.svg)](./LICENSE)

## Requirements ##

You need a working ZFS setup with a few patches, in the following order.

1. stable api https://github.com/zfsonlinux/zfs/pull/3907
1. patch on top of stable api https://github.com/ClusterHQ/zfs/pull/13.patch

Generally you need root privileges to use anything zfs related.

## Status ##

![status=alpha](https://img.shields.io/badge/status-alpha-orange.svg)

We have implemented most of the features we need.
If any features are wanted please open an issue or (better yet) pull request.
The beta TODO list is more or less:

- [ ] stabilize `zfs_receive`
- [ ] avoid `zfs_cmd_t` in its current form
- [ ] implement usable public api akin to [go-zfs](https://github.com/mistifyio/go-zfs)
- [ ] find a better name

# Contributing #

See the [contributing guidelines](./CONTRIBUTING.md)

# cgo #
All the nv list handling is fully implemented in go.
The only use of cgo is a wrapper that actually calls `ioctl` with the `zfs_cmd_t`, this may be changed in the future.
