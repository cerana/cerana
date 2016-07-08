#!/bin/bash

# shellcheck disable=SC1091
source /tmp/cerana-bootcfg

# don't configure networking in rescue mode
[[ -n $CERANA_RESCUE ]] && exit 0

# Create associative arrays of interfaces and their MAC addresses
declare -A MAC_TO_IFACE
declare -A IFACE_TO_MAC
function collect_addresses() {
    local name address
    for device in /sys/class/net/e*; do
        name=${device#/sys/class/net/}
        read -r address <"${device}/address"
        MAC_TO_IFACE[$address]=$name
        IFACE_TO_MAC[$name]=$address
    done
}

# Collect them already. Other functions need them.
collect_addresses

function fail_exit() {
    local return
    return=$1
    shift
    echo "$*" >&2
    exit "$return"
}

## is_valid_ip <ip address>
##
## Takes a dotted-quad format IP address and returns based on its validity
function is_valid_ip() {

    local ip=$1
    local stat=1

    if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        OIFS=$IFS
        IFS='.'
        ip=($ip)
        IFS=$OIFS
        [[ ${ip[0]} -le 255 && ${ip[1]} -le 255 \
            && ${ip[2]} -le 255 && ${ip[3]} -le 255 ]]
        stat=$?
    fi
    return $stat
}

## is_valid_prefix <cidr prefix>
##
## Takes a CIDR prefix and makes sure that it is greater than 7 and less than 33
function is_valid_prefix() {
    local prefix=$1
    [[ -n $prefix ]] && ((8 <= prefix && prefix <= 32))
    return $?
}

function query_interface() {
    local response
    response="not-an-interface"
    # If there's only one interface, use it. No need to prompt user.
    [[ "1" == "${#IFACE_TO_MAC[@]}" ]] && response="${!IFACE_TO_MAC[*]}"
    echo
    echo "Detected interfaces:"
    echo
    for iface in "${!IFACE_TO_MAC[@]}"; do
        echo "$iface		${IFACE_TO_MAC[$iface]}"
    done

    while [[ -z "${IFACE_TO_MAC[$response]}" ]]; do
        echo
        echo "Please choose an interface for management network DHCP."
        echo -n "Selection: "
        read -r response
    done
    CERANA_MGMT_MAC=${IFACE_TO_MAC[$response]}
}

function query_ip() {
    local answer
    local ip
    local prefix

    answer=${CERANA_MGMT_IP}
    ip=${answer%/*}
    prefix=${answer#*/}

    while [[ -z "$answer" ]] \
        || [[ -z "$prefix" ]] \
        || ! is_valid_ip "$ip" \
        || ! is_valid_prefix "$prefix"; do
        echo
        echo "Please specify the IP address and netmask of this node in the form a.b.c.d/n :"
        echo
        echo -n "> "
        read -r answer
        ip=${answer%/*}
        prefix=${answer#*/}
    done

    CERANA_MGMT_IP=$answer
}

function config_mgmt_dhcp() {
    [[ -n ${CERANA_MGMT_MAC} ]] \
        && [[ -n ${MAC_TO_IFACE[${CERANA_MGMT_MAC}]} ]] \
        || return 1
    echo -e "[Match]\nMACAddress=${CERANA_MGMT_MAC}\n\n[Network]\nDHCP=yes\n" >/data/config/network/mgmt.network
    # If we're on SmartOS, fall back to using a MAC based ClientIdentifier to avoid being blocked by antispoof
    lshw | grep -iq joyent \
        && echo -e "\n[DHCP]\nClientIdentifier=mac\n" >>/data/config/network/mgmt.network
    # If grep fails we still want to return true otherwise the service can fail
    true
}

function config_mgmt_static() {
    [[ -n ${CERANA_MGMT_MAC} ]] \
        && [[ -n ${MAC_TO_IFACE[${CERANA_MGMT_MAC}]} ]] \
        && [[ -n ${CERANA_MGMT_IP} ]] \
        || return 1
    echo -e "[Match]\nMACAddress=${CERANA_MGMT_MAC}\n\n[Network]\nAddress=${CERANA_MGMT_IP}" >/data/config/network/mgmt.network
}

function drop_consul_config() {
    # FIXME Need to implement joining a cluster as a "client" agent
    local CONFIG comma
    CONFIG=/data/config/consul.json

    echo '{' >$CONFIG
    echo '"server": true,' >>$CONFIG
    case $1 in
        'bootstrap')
            echo '"bootstrap": true,' >>$CONFIG
            ;;
        'join')
            echo '"start_join": [' >>$CONFIG
            for server in ${CERANA_CLUSTER_IPS//,/ }; do
                [[ -n $comma ]] && echo -n ', ' >>$CONFIG || comma=1 # json is terrible
                echo "\"$server\"" >>$CONFIG
            done
            echo '],' >>$CONFIG
            ;;
        *)
            return 1
            ;;
    esac
    echo '"data_dir": "/data/config/consul/"' >>$CONFIG
    echo '}' >>$CONFIG
}

export_config() {
    declare | grep ^CERANA >/data/config/cerana-bootcfg
}

enable_layer2() {
    mkdir -p /data/services/cerana.target.wants/
    ln -s /etc/systemd/system/ceranaLayer2.target /data/services/cerana.target.wants/
}

## Main
collect_addresses

if [[ -n $CERANA_CLUSTER_BOOTSTRAP ]]; then
    # We're bootstrapping a layer 2 cluster
    # 1. Which MAC address are we using
    config_mgmt_dhcp \
        || query_interface
    # 2. What IP range are we using?
    query_ip
    #3. Set the IP statically
    config_mgmt_static

    drop_consul_config bootstrap \
        || fail_exit $? "Cluster configuration failed"

    # remove the bootstrap flag for future boots
    unset CERANA_CLUSTER_BOOTSTRAP

    export_config

elif [[ -n $CERANA_CLUSTER_IPS ]]; then
    # We're joining a layer 2 cluster
    config_mgmt_dhcp \
        || { query_interface \
            && query_ip \
            && config_mgmt_static; }
    drop_consul_config join \
        || fail_exit $? "Cluster joining failed"
    export_config

else
    # We're a standalone node
    # If we were given a MAC address to use on the kernel command line, use it
    # Otherwise, prompt for which device to use and ask about DHCP or static config
    config_mgmt_dhcp && exit 0

    #FIXME Only do DHCP for now. Static IP for standalone will come later.
    query_interface
    export_config
    config_mgmt_dhcp
fi
