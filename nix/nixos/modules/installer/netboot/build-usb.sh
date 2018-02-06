#!/usr/bin/env bash

# useful debugging settings
#trap '{ type -p -- "${BASH_COMMAND%% *}" >&3; } 3>&2 2> /dev/null' DEBUG
#set -o functrace -o xtrace

# tell me what you're doing
set -o xtrace

# use binaries from Nix, not from host system
PACKAGES="grub bash multipath-tools coreutils dosfstools gnufdisk utillinux"

# make sure packages are already in the Nix store before running as root.
nix-shell -p ${PACKAGES} --run true

# rerun this as root
[[ $USER == root ]] || exec sudo NIX_PATH="${NIX_PATH}" $(which nix-shell) --run $0 -p ${PACKAGES}

# find grub package for stage{1,2}
GRUB=$(which grub)
GRUB=${GRUB/bin*grub}

DISK=/var/tmp/cerana.img
MOUNT=/mnt

# clean up from any possible failed runs
kpartx -d ${DISK}
rm ${DISK}

mkdir -p ${MOUNT}
dd if=/dev/zero of=${DISK} bs=1M seek=1023 count=1
echo -en "n\n\n\n\n\nt\nb\np\nw\n" | fdisk ${DISK}

FILESYSTEM="/dev/mapper/$(kpartx -av ${DISK} | tee /dev/stderr | grep p1 | sed 's|p1 .*|p1|;s|.* ||')"
sleep 3

#Create filesystem, copy files
mkfs.vfat -n CERANA ${FILESYSTEM}
mount ${FILESYSTEM} ${MOUNT} || exit 1
mkdir -p ${MOUNT}/boot/grub/
cp ${GRUB}/lib/grub/i386-pc/stage? ${MOUNT}/boot/grub/
cp result/menu.lst ${MOUNT}/boot/grub
cp result/{bzImage,initrd} ${MOUNT}
sync
umount ${MOUNT}

# install grub
${GRUB}/bin/grub --batch <<____ENDOFGRUBCOMMANDS
device (hd0) $DISK
root (hd0,0)
setup (hd0)
____ENDOFGRUBCOMMANDS

# clean up loop devices and fix permissions
kpartx -dv ${DISK}
chown ${SUDO_USER} ${DISK}
