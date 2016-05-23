#!/bin/bash

CERANA_IFSTATE=/tmp/.ifstate
CERANA_CONFIG=/tmp/cerana-config
CERANA_BOOTCFG=/tmp/cerana-bootcfg
CERANA_MAC=/etc/sysconfig/ovsbridge
PROC_CMDLINE=/proc/cmdline
ETHER_DEVS=(`ls /sys/class/net | egrep -v '^lo$'`)

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
    local cmdline=$(cat $PROC_CMDLINE)
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

# Act on any IP address being passed to us via kernel boot arguments.
# By default, configure the first enummerated interface on the system with it.
function configure_net_manual() {
    if [ -n "$CERANA_KOPT_MGMT_IFACE" ]; then
        local iface="$CERANA_KOPT_MGMT_IFACE"
    else
        local iface="${ETHER_DEVS[0]}"
    fi

    echo "Manually configuring $iface with $CERANA_KOPT_MGMT_IP..."
    /sbin/ip link set $iface up
    /sbin/ip addr add $CERANA_KOPT_MGMT_IP dev $iface
    echo "IFTYPE=MANUAL" >> $CERANA_IFSTATE
    echo "IFACE=$iface" >> $CERANA_IFSTATE
    echo "IP=$CERANA_KOPT_MGMT_IP" >> $CERANA_IFSTATE

    if [ -n "$CERANA_KOPT_MGMT_GW" ]; then
        echo "Manually adding $CERANA_KOPT_MGMT_GW as default route"
        /sbin/ip route add default via $CERANA_KOPT_MGMT_GW
        echo "GW=$CERANA_KOPT_MGMT_GW" >> $CERANA_IFSTATE
    fi

    save_mac_for_bridge $iface
}

# Cycle through all known ethernet interfaces until we find one which
# responds to DHCP requests.
function configure_net_dhcp() {
    set -x
    echo "Probing for DHCP"
    /sbin/dhclient -v -e CERANA_IFSTATE=$CERANA_IFSTATE -1 ${ETHER_DEVS[*]} &&
        echo "IFTYPE=DHCP" >> $CERANA_IFSTATE
}

# Unconfigure and otherwise clean up any interfaces we may have configured
function unconfigure_net_iface() {
    if [ -f $CERANA_IFSTATE ]; then
        . $CERANA_IFSTATE
    else
        return 1
    fi

    if [ "$IFTYPE" != "DHCP" ]; then
        return
    fi

    echo "Releasing DHCP lease on $IFACE..."
    /sbin/dhclient -v -e CERANA_IFSTATE=$CERANA_IFSTATE -r ${ETHER_DEVS[*]}
    /sbin/ip addr del $IP dev $IFACE
    /sbin/ip link set $IFACE up

    save_mac_for_bridge $IFACE
}

function save_mac_for_bridge() {
    local mac=$(cat /sys/class/net/$1/address)
    ! grep -sq MACAddress $CERANA__KOPT_MGMT_MAC &&
        echo "MACAddress=$mac" >> $CERANA_KOPT_MGMT__MAC &&
        echo "Saving MAC address for bridge $mac"
}

function get_cerana_config() {
    if [ -f $CERANA_IFSTATE ]; then
        . $CERANA_IFSTATE
    else
        return 1
    fi

    if [[ -z $DNS ]]; then
        # not configured via dhcp
        return 1
    fi

    local resp=''
    for dns in ${DNS[*]}; do
        srv=$(dig +noall +answer +additional SRV ipxe.services.lochness.local @$dns) || continue
        resp="$srv"
        break
    done

    if [[ -z $resp ]]; then
        echo "could not resolv ipxe service"
        return 1
    fi

    local addr=$(echo $resp | awk '/\sA\s/ {print $NF}')
    local port=$(echo $resp | awk '/\sSRV\s/ {print $7}')
    local ip=${IP%%/*}
    curl http://$addr:$port/config/$ip > $CERANA_CONFIG

    if [ $? -ne 0 ]; then
        echo "FATAL: Could not get Cerana configuration for $ip!"
        return 1
    fi
}

## Network initialization main. Under 'start' we either manually configure a IP address
## on the first interface if we are given one in the kernel boot arguments.
## Otherwise we probe each interface for a DHCP response until we fine one that does.
case "$1" in
    'start')
        # Initialize our own interface state file
        cp /dev/null $CERANA_IFSTATE

        parse_boot_args

	# Test if the cerana.install option was specified. Start install
	# script to quiz the user for basic IP information
	if [ $CERANA_KOPT_INSTALL ]; then
		/etc/pre-init.d/install.sh
		break
	fi

        if [ -n "$CERANA_KOPT_MGMT_IP" ]; then
            configure_net_manual
        else
            configure_net_dhcp
            get_cerana_config
        fi
        ;;
    'stop')
        unconfigure_net_iface
        ;;
    *)
        echo "Usage: $0 {start|stop}"
        ;;
esac

# vim:set ts=4 sw=4 et:
