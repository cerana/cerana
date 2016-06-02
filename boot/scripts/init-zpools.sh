#!/bin/bash

source /tmp/cerana-bootcfg

## The name of the zpool we're concerned with. Everything revolves around this
## being configured and present. Set name accordingly.
ZPOOL=data
FILESYSTEMS="datasets datasets/ro datasets/rw running-clones config services logs platform"

# Create the expected filesystems on the configured zpool
configure_filesystems()
{
	# Do we have the zpool?
	zpool list $ZPOOL > /dev/null 2>&1
	if [ $? -ne 0 ]; then
	    echo "ERROR: $ZPOOL zpool is not present!"
	    return 1
	fi

	# Create base ZFS filesystems and set their properies
	for fs in $FILESYSTEMS; do
		zfs list $ZPOOL/$fs > /dev/null 2>&1 || zfs create $ZPOOL/$fs
	done

	local compression=$(zfs get compression -Ho value $ZPOOL)
	if [ "$compression" = "off" ]; then
		zfs set compression=lz4 $ZPOOL
	fi

	local quota=$(zfs get quota -Ho value $ZPOOL/private)
	if [ "$quota" = "none" ]; then
		zfs set quota=32G $ZPOOL/private
	fi
}

# Gather a list of disk devices into an array
get_disklist()
{
	# Get a basic list of disk devices in the format:
	# disk sda     16G VBOX HARDDISK
	# disk sdb     16G VBOX HARDDISK
	# disk sdc     45G VBOX HARDDISK
	# disk sdd     10G VBOX HARDDISK
	DISKLIST=`lsblk --ascii --noheadings \
		  --output type,name,size,model --nodeps | grep ^disk`

	# Scrape the above into a list of just device nodes
	DISKDEVS=`echo "$DISKLIST" | awk '{print "/dev/"$2}'`

	readarray DISKARRAY <<< "$DISKDEVS"
}

# Prompt the user for the type of zpool they desire. Offer a command shell
# to allow manual configuration using the zpool command directly.
prompt_user()
{
	echo
	echo "No existing \"$ZPOOL\" zpool detected."
	echo "In order to proceed, you must configure the underlying ZFS"
	echo "storage for this node. The following disks have been detected:"
	echo
	echo "$DISKLIST"
	echo
	echo "The available automatic configuration options are:"
	echo
	if [ ${#DISKARRAY[@]} -eq 0 ]; then
		echo "ERROR: This server has no disks!"
		exit
	elif [ ${#DISKARRAY[@]} -eq 1 ]; then
		echo " stripe (${#DISKARRAY[@]} disk) *"
		AUTO="stripe"
	elif [ ${#DISKARRAY[@]} -eq 2 ]; then
		echo " stripe (${#DISKARRAY[@]} disks) *"
		echo " mirror (${#DISKARRAY[@]} disks)"
		AUTO="mirror"
	elif [ ${#DISKARRAY[@]} -eq 3 ]; then
		echo " stripe (${#DISKARRAY[@]} disks) *"
		echo " raidz (${#DISKARRAY[@]} disks)"
		AUTO="raidz"
	elif [ ${#DISKARRAY[@]} -eq 4 ]; then
		echo " stripe (${#DISKARRAY[@]} disks) *"
		echo " raidz (${#DISKARRAY[@]} disks)"
		echo " raidz2 (${#DISKARRAY[@]} disks)"
		AUTO="raidz"
	elif [ ${#DISKARRAY[@]} -gt 4 ]; then
		echo " stripe (${#DISKARRAY[@]} disks) *"
		echo " raidz (${#DISKARRAY[@]} disks)"
		echo " raidz2 (${#DISKARRAY[@]} disks)"
		echo " raidz3 (${#DISKARRAY[@]} disks)"
		AUTO="raidz2"
	fi
	echo
	echo "* = not suggested for production purposes"
	echo
	echo "You may also type \"manual\" to enter a shell to perform a manual"
	echo "configuration of the \"${ZPOOL}\" zpool using command-line tools."
	echo "This is preferable if an exact ZFS pool configuration is desired,"
	echo "such as one that contains hot spares, or slog/L2ARC devices."
	echo
	echo "Please reference the ZFSonLinux FAQ:"
	echo "http://zfsonlinux.org/faq.html"
	echo

	answer=${CERANA_KOPT_ZFS_CONFIG}
	while [ ! $answer ]; do
		echo -n "Selection: "
		read answer
	done

	case $answer in
	'auto')
		answer=${AUTO}
		;&
	'raidz')
		;&
	'raidz2')
		;&
	'raidz3')
		;&
	'mirror')
		;&
	'stripe')
		POOLTYPE="$answer"
		echo "Using all disks to create a $answer zpool..."
		;;
	'manual')
		POOLTYPE="manual"
		echo "Entering shell to perform manual pool configuration."
		echo "You must configure a zpool with the name of \"$ZPOOL\"."
		;;
	*)
		return 1
		;;
	esac
}

# Create the zpool using user input
configure_zpool()
{
	# If manual pool creation was selected, run a sub-shell and check
	# to make sure the pool was actually created after the sub-shell
	# was exited. If not, run the shell again.
	if [ "$POOLTYPE" = "manual" ]; then
		# Execute shell for manual zpool config; then create zfs FSs
		echo "When finished, type \"exit\" or Ctrl-D"
		echo "The option \"-o cachefile=none\" MUST be specified"
		PS1='zpool-config> ' /bin/bash
		return
	fi

	# Stripes are easy
	if [ "$POOLTYPE" == "stripe" ]; then
		ZPOOLARGS="${DISKARRAY[*]}"
	fi

	# Mirrors are a little tricky if we want to support more than a single
	# mirrored pair
	#
	# TODO: Works for 1 pair of disks. If we have an even number of disks,
	# 2 pairs or more, we should do the math and go through the effort to
	# make that an option.
	if [ "$POOLTYPE" == "mirror" ]; then
		ZPOOLARGS="mirror ${DISKARRAY[*]}"
	fi

	# RAIDZ(n)
	if [[ "$POOLTYPE" == "raidz"* ]]; then
		ZPOOLARGS="$POOLTYPE ${DISKARRAY[*]}"
	fi

	# Be sure the drives are starting fresh. Zero out any existing partition
	for d in "${DISKARRAY[@]}"; do
	    echo "Clearing $d"
	    sgdisk -Z $d > /dev/null 2>&1
	done

	# Block until Linux figures itself out w.r.t. paritions
	udevadm settle

	# Now create the zpool
	zpool create -f \
		-o cachefile=none \
		$ZPOOL $ZPOOLARGS

	if [ $? -ne 0 ]; then
	    return 1
	fi

	# Pause until the zpool is imported and mounted
	while [ ! -d /$ZPOOL ]
	do
		sleep 2
	done

}

## Main logic

POOLTYPE=""

udevadm settle
zpool import -f $ZPOOL > /dev/null 2>&1

# We did not detect a zpool by the name of $ZPOOL, so start the process to
# create one.
if [ $? -ne 0 ]; then
	while [ ! -d /$ZPOOL ]; do
		get_disklist
		while [ ! "$POOLTYPE" ]; do
			prompt_user
		done
		configure_zpool
	done
fi

# Check to make sure our required filesystems are on the zpool, and create
# them if any aren't.
configure_filesystems
