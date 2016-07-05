#!/bin/bash

# shellcheck disable=SC1091
source /tmp/cerana-bootcfg

if [[ -L /dev/disk/by-label/CERANA ]]; then
    # The CD/USB is in the drive
    mkdir -p /mnt
    mount /dev/disk/by-label/CERANA /mnt
    mkdir -p /data/platform/bootcd
    rsync -a /mnt/ /data/platform/bootcd/
    umount /mnt
    rm -f /data/platform/current
    ln -s bootcd /data/platform/current
    # No physical media, most likely PXE booted
    # FIXME
    # Check to see if we already have latest media on disk
    # If not, download it and update /data/platform/current symlink
fi
