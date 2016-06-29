#!/bin/bash

CERANA_BOOTCFG=/tmp/cerana-bootcfg
CERANA_PREFIX="cerana"
CMDLINE=$(cat /proc/cmdline)

# Parse the kernel boot arguments and look for our arguments and
# pick out their values.
function parse_boot_args() {
    >$CERANA_BOOTCFG

    [[ ${CMDLINE} =~ ${CERANA_PREFIX} ]] || return

    for kopt in $CMDLINE; do
        IFS='=' read -r k v <<<"$kopt"
        case $k in
            $CERANA_PREFIX.zfs_config)
                echo "CERANA_ZFS_CONFIG=$v" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.cluster_ips)
                echo "CERANA_CLUSTER_IPS=$v" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.cluster_bootstrap)
                # No argument expected. If present, just set to 1.
                echo "CERANA_CLUSTER_BOOTSTRAP=1" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.rescue)
                # No argument expected. If present, just set to 1.
                echo "CERANA_RESCUE=1" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.mgmt_mac)
                echo "CERANA_MGMT_MAC=$v" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.mgmt_ip)
                echo "CERANA_MGMT_IP=$v" >>$CERANA_BOOTCFG
                ;;
            $CERANA_PREFIX.mgmt_gw)
                echo "CERANA_MGMT_GW=$v" >>$CERANA_BOOTCFG
                ;;
        esac
    done
}

#if manually invoked with argument "test" do a quick sanity check to stdout
if [[ "test" == "$1" ]]; then
    CERANA_BOOTCFG=/dev/stdout
    CMDLINE="cerana.zfs_config=auto cerana.cluster_ips=10.2.3.4,10.2.3.5 cerana.cluster_bootstrap cerana.rescue cerana.mgmt_mac=00:00:00:00 cerana.mgmt_ip=10.2.3.6/24 cerana.mgmt_gw=10.2.3.1"
fi

parse_boot_args
