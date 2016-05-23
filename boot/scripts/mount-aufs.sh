#!/bin/bash
set -e

declare -a OVERLAYS=(
    etc
    var
    root
    home
)

# Create AUFS overlay mountpoints under /mistify/private
mount_aufs () {
	for D in ${OVERLAYS[@]}; do
		mode=`stat -c %a /$D`

		[ -d /mistify/private/$D ] || mkdir /mistify/private/$D
		chmod $mode /mistify/private/$D

		echo "Mounting AUFS filesystem /$D"
		mount -t aufs -o br:/mistify/private/$D:/$D=ro none $D
	done
}

mount_aufs

exit 0
