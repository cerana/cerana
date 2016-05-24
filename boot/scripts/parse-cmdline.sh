#!/bin/bash

CERANA_BOOTCFG=/tmp/cerana-bootcfg
CERANA_KOPT_PREFIX="cerana"

CERANA_KOPT_ZFS_CONFIG=""
CERANA_KOPT_CLUSTER_IPS=""
CERANA_KOPT_CLUSTER_BOOTSTRAP=""
CERANA_KOPT_RESCUE=""
CERANA_KOPT_MGMT_MAC=""
CERANA_KOPT_MGMT_IP=""
CERANA_KOPT_MGMT_GW=""
CERANA_KOPT_MGMT_IFACE=""

# Parse the kernel boot arguments and look for our arguments and
# pick out their values.
function parse_boot_args() {
    local cmdline=$(cat /proc/cmdline)
    grep -q $CERANA_KOPT_PREFIX <<<$cmdline || return

    cp /dev/null $CERANA_BOOTCFG

    for kopt in $cmdline; do
        k=`echo $kopt | awk -F= '{print $1}'`
        v=`echo $kopt | awk -F= '{print $2}'`
        case $k in
            $CERANA_KOPT_PREFIX.zfs_config)
                CERANA_KOPT_ZFS_CONFIG="$v"
		echo "CERANA_KOPT_ZFS_CONFIG=1" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.cluster_ips)
                CERANA_KOPT_CLUSTER_IPS="$v"
		echo "CERANA_KOPT_CLUSTER_IPS=1" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.cluster_bootstrap)
                CERANA_KOPT_CLUSTER_BOOTSTRAP="$v"
		echo "CERANA_KOPT_CLUSTER_BOOTSTRAP=1" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.rescue)
		# No argument expected. If present, just set to 1.
		echo "CERANA_KOPT_RESCUE=1" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.install)
		# No argument expected. If present, just set to 1.
		echo "CERANA_KOPT_INSTALL=1" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.mgmt_mac)
                CERANA_KOPT_MGMT_MAC="$v"
		echo "CERANA_KOPT_MGMT_MAC=$v" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.mgmt_ip)
                CERANA_KOPT_MGMT_IP="$v"
		echo "CERANA_KOPT_MGMT_IP=$v" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.mgmt_gw)
                CERANA_KOPT_MGMT_GW="$v"
		echo "CERANA_KOPT_MGMT_GW=$v" >> $CERANA_BOOTCFG
                ;;
            $CERANA_KOPT_PREFIX.mgmt_iface)
                CERANA_KOPT_MGMT_IFACE="$v"
		echo "CERANA_KOPT_MGMT_IFACE=$v" >> $CERANA_BOOTCFG
                ;;
        esac
    done
}

parse_boot_args
