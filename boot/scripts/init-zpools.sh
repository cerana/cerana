#!/bin/bash

# shellcheck disable=SC1091
source /tmp/cerana-bootcfg
[[ -n $CERANA_RESCUE ]] && exit 0

## The name of the zpool we're concerned with. Everything revolves around this
## being configured and present. Set name accordingly.
ZPOOL=data
FILESYSTEMS="datasets datasets/ro datasets/rw running-clones config services logs platform"

# Create the expected filesystems on the configured zpool
configure_filesystems() {

    # Do we have the zpool?
    zpool list $ZPOOL >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "ERROR: \"$ZPOOL\" pool is not present!"
        return 1
    fi

    # Create base ZFS filesystems and set their properies
    for fs in $FILESYSTEMS; do
        zfs list "$ZPOOL/$fs" >/dev/null 2>&1 || zfs create "$ZPOOL/$fs"
    done

    if [[ "off" = $(zfs get compression -Ho value $ZPOOL) ]]; then
        zfs set compression=lz4 $ZPOOL
    fi
}

# Gather a list of disk devices into an array
get_disklist() {

    # Get a basic list of disk devices in the format:
    # disk sda     16G VBOX HARDDISK
    # disk sdb     16G VBOX HARDDISK
    # disk sdc     45G VBOX HARDDISK
    # disk sdd     10G VBOX HARDDISK
    DISKLIST=$(lsblk --ascii --noheadings \
        --output type,name,size,model --nodeps | grep ^disk)

    # Scrape the above into a list of just device nodes
    DISKDEVS=$(echo "$DISKLIST" | awk '{print "/dev/"$2}')

    readarray DISKARRAY <<<"$DISKDEVS"
}

# Prompt the user for the type of pool they desire. Offer a command shell
# to allow manual configuration using the zpool command directly.
prompt_user() {

    echo
    echo "No existing \"$ZPOOL\" pool detected."
    echo "In order to proceed, you must configure the underlying ZFS"
    echo "storage for this node. The following disks have been detected:"
    echo
    echo "$DISKLIST"
    echo
    echo "The available automatic configuration options are:"
    echo
    if [[ ${#DISKARRAY[@]} -eq 0 ]]; then
        echo "ERROR: This server has no disks!"
        exit
    elif [[ ${#DISKARRAY[@]} -eq 1 ]]; then
        echo " stripe (${#DISKARRAY[@]} disk) * (auto)"
        AUTO="stripe"
    elif [[ ${#DISKARRAY[@]} -eq 2 ]]; then
        echo " stripe (${#DISKARRAY[@]} disks) *"
        echo " mirror (${#DISKARRAY[@]} disks) (auto)"
        AUTO="mirror"
    elif [[ ${#DISKARRAY[@]} -eq 3 ]]; then
        echo " stripe (${#DISKARRAY[@]} disks) *"
        echo " raidz (${#DISKARRAY[@]} disks) (auto)"
        AUTO="raidz"
    elif [[ ${#DISKARRAY[@]} -eq 4 ]]; then
        echo " stripe (${#DISKARRAY[@]} disks) *"
        echo " raidz (${#DISKARRAY[@]} disks) (auto)"
        echo " raidz2 (${#DISKARRAY[@]} disks)"
        AUTO="raidz"
    elif [[ ${#DISKARRAY[@]} -gt 4 ]]; then
        echo " stripe (${#DISKARRAY[@]} disks) *"
        echo " raidz (${#DISKARRAY[@]} disks)"
        echo " raidz2 (${#DISKARRAY[@]} disks) (auto)"
        echo " raidz3 (${#DISKARRAY[@]} disks)"
        AUTO="raidz2"
    fi
    echo
    echo "* = not suggested for production purposes"
    echo "(auto) = configuration used if you type \"auto\""
    echo
    echo "You may type \"auto\" and let the system choose a layout for you"
    echo "You may also type \"manual\" to enter a shell to perform a manual"
    echo "configuration of the \"${ZPOOL}\" pool using command-line tools."
    echo "This is preferable if an exact ZFS pool configuration is desired,"
    echo "such as one that contains hot spares, or slog/L2ARC devices."
    echo
    echo "Please reference the ZFSonLinux FAQ:"
    echo "http://zfsonlinux.org/faq.html"
    echo

    answer=${CERANA_ZFS_CONFIG}
    while [[ ! $answer ]]; do
        echo -n "Selection: "
        read -r answer
    done

    case $answer in
        'auto')
            answer=${AUTO}
            ;&
        'raidz') ;&
        'raidz2') ;&
        'raidz3') ;&
        'mirror') ;&
        'stripe')
            POOLTYPE="$answer"
            echo "Using all disks to create a $answer pool..."
            ;;
        'manual')
            POOLTYPE="manual"
            echo "Entering shell to perform manual pool configuration."
            echo "You must configure a pool with the name of \"$ZPOOL\"."
            ;;
        *)
            return 1
            ;;
    esac
}

# add a partition for and install GRUB2 to pool disks
install_grub() {
    local node major minor
    [[ -n ${INSTALL_GRUB_PLEASE} ]] || return
    # shellcheck disable=SC2068
    for disk in ${DISKARRAY[@]}; do
        echo -n "Adding BIOS boot partition to $disk"
        # /dev isn't mounted yet (?!) time for mknod!
        node=${disk#/dev}
        IFS=: read -r major minor <"/sys/block/$node/dev"
        mknod "$node" b "$major" "$minor"
        echo -e "n\n2\n34\n2047\nt\n2\n4\nw\n" | fdisk "$node" &>/dev/null \
            && echo "  done"
        echo -n "Installing GRUB boot block on $disk"
        grub-install --modules=zfs "$node" &>/dev/null \
            && echo "  done"
        rm "$node"
    done
}

# Minimal workable GRUB2 from ZFS (uses serial port)
create_grub_config() {
    [[ -n ${INSTALL_GRUB_PLEASE} ]] || return
    [[ -f /data/boot/grub/grub.cfg ]] && return
    mkdir -p /data/boot/grub
    ln -s /data/boot /boot
    cat >/data/boot/grub/grub.cfg <<'EOF'
serial --unit=0 --speed=115200
terminal_input serial
terminal_output serial

set default 0
set timeout 10
set color_normal=white/black
set color_highlight=black/white

search --set=poolname --label data

menuentry "Boot CeranaOS from Disk" {
  linux ($poolname)/platform/@//current/bzImage loglevel=4 console=ttyS0
  initrd ($poolname)/platform/@//current/initrd
}

menuentry "Boot CeranaOS Rescue Mode" {
  linux ($poolname)/platform/@//current/bzImage loglevel=4 console=ttyS0 cerana.rescue
  initrd ($poolname)/platform/@//current/initrd
}
EOF
}

# Create the zpool using user input
configure_pool() {

    # If manual pool creation was selected, run a sub-shell and check
    # to make sure the pool was actually created after the sub-shell
    # was exited. If not, run the shell again.
    if [[ "$POOLTYPE" = "manual" ]]; then
        # Execute shell for manual zpool config; then create zfs FSs
        echo "When finished, type \"exit\" or Ctrl-D"
        echo "The option \"-o cachefile=none\" MUST be specified"
        PS1='zpool-config> ' bash
        return
    fi

    # Stripes are easy
    if [[ "$POOLTYPE" == "stripe" ]]; then
        ZPOOLARGS="${DISKARRAY[*]}"
    fi

    # Mirrors are a little tricky if we want to support more than a single
    # mirrored pair
    #
    # TODO: Works for 1 pair of disks. If we have an even number of disks,
    # 2 pairs or more, we should do the math and go through the effort to
    # make that an option.
    if [[ "$POOLTYPE" == "mirror" ]]; then
        ZPOOLARGS="mirror ${DISKARRAY[*]}"
    fi

    # RAIDZ(n)
    if [[ "$POOLTYPE" == "raidz"* ]]; then
        ZPOOLARGS="$POOLTYPE ${DISKARRAY[*]}"
    fi

    # Be sure the drives are starting fresh. Zero out any existing partition
    # shellcheck disable=SC2068
    for disk in ${DISKARRAY[@]}; do
        echo -n "Clearing $disk"
        sgdisk -Z "$disk" &>/dev/null \
            && echo "  done"
    done

    # Flag for installing GRUB onto disks
    INSTALL_GRUB_PLEASE=1

    # Now create the zpool
    # shellcheck disable=SC2086
    zpool create -f \
        -o cachefile=none \
        $ZPOOL $ZPOOLARGS

    if [[ $? -ne 0 ]]; then
        return 1
    fi

    # Pause until the zpool is imported and mounted
    while [[ ! -d /$ZPOOL ]]; do
        sleep 2
    done

}

## Main logic

POOLTYPE=""

udevadm settle
zpool import -f $ZPOOL >/dev/null 2>&1

# We did not detect a zpool by the name of $ZPOOL, so start the process to
# create one.
if [[ $? -ne 0 ]]; then
    while [[ ! -d /$ZPOOL ]]; do
        get_disklist
        while [[ ! "$POOLTYPE" ]]; do
            prompt_user
        done
        configure_pool
    done
fi

# Check to make sure our required filesystems are on the zpool, and create
# them if any aren't.
configure_filesystems

# Add a grub config if we don't have one
create_grub_config

# Install GRUB if this is a fresh pool
install_grub

# Attempt to read in any on-disk config from a previous boot and update with info from kernel command line
if [[ -f /data/config/cerana-bootcfg ]]; then
    cp /data/config/cerana-bootcfg /data/config/cerana-previous-bootcfg
    # shellcheck disable=SC1091
    source /data/config/cerana-bootcfg
fi

# Store current config (kernel command line arguments always override what was found on disk)
# shellcheck disable=SC1091
source /tmp/cerana-bootcfg
declare | grep ^CERANA >/data/config/cerana-bootcfg
rm /tmp/cerana-bootcfg
ln -s /data/config/cerana-bootcfg /tmp/cerana-bootcfg

# Link in network config directory for systemd-networkd
mkdir -p /data/config/network
ln -s /data/config/network /run/systemd/network

mkdir /etc/systemd-mutable
ln -s /data/services /etc/systemd-mutable/system

# create the mutable cerana.target.wants
mkdir -p /data/services/cerana.target.wants

#if we're a cluster node, create this symlink for layer 2 services
[[ -n $CERANA_CLUSTER_BOOTSTRAP ]] \
    || [[ -n $CERANA_CLUSTER_IPS ]] \
    && ln -s /etc/systemd/system/ceranaLayer2.target /data/services/cerana.target.wants/

# load in unit files that already exist in the pool
systemctl daemon-reload

# start up the full cerana target
systemctl --no-block start cerana.target
