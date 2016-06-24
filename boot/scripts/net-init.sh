#!/bin/bash

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
        read address <${device}/address
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
    exit $return
}

## in_array ARRAY ITEM
##
## Tests whether ITEM is contained inside array ARRAY
function in_array() {
    local haystack=${1}[@]
    local needle=${2}
    for i in ${!haystack}; do
        if [[ ${i} == ${needle} ]]; then
            return 0
        fi
    done
    return 1
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
    local stat=1

    if [[ $prefix =~ ^[0-9]{1,2}$ ]]; then
        [[ $prefix -le 32 || $prefix -ge 8 ]]
        stat=$?
    fi

    return $stat
}

## iface_list
##
## Generates an array IFACES with a list of physical ethernet interfaces
function iface_list() {
    pushd .
    cd /sys/class/net
    IFACES=(e*)
    popd
}

function query_interface() {
    local response
    response="not-an-interface"
    echo
    echo "Detected interfaces:"
    echo
    for iface in ${!IFACE_TO_MAC[@]}; do
        echo "$iface		${IFACE_TO_MAC[$iface]}"
    done

    while [[ -z ${IFACE_TO_MAC[$response]} ]]; do
        echo
        echo "Please choose an interface for management network DHCP."
        echo -n "Selection: "
        read response
    done
    CERANA_MGMT_MAC=${IFACE_TO_MAC[$response]}
}

function query_ip() {
    local answer
    local ip
    local prefix

    answer=${CERANA_MGMT_IP}

    ip=$(awk -F/ '{print $1}' <<<$answer)
    prefix=$(awk -F/ '{print $2}' <<<$answer)

    while [[ ! $answer ]] \
        || [[ ! $prefix ]] \
        || ! is_valid_ip $ip \
        || ! is_valid_prefix $prefix; do
        echo
        echo "Please specify the IP address and netmask of this node in the form a.b.c.d/n :"
        echo
        echo -n "> "
        read answer

        ip=$(awk -F/ '{print $1}' <<<$answer)
        prefix=$(awk -F/ '{print $2}' <<<$answer)
    done

    CERANA_IP=$ip
    CERANA_NETMASK=$prefix
}

function query_gateway() {
    local answer

    echo
    echo "Please specify the default gateway IP:"
    echo
    echo -n "> "
    read answer

    while [ ! $answer ] || ! is_valid_ip $answer; do
        echo "You must provide a valid IP address!"
        echo -n "> "
        read answer
    done

    CERANA_GW=$answer
}

function config_mgmt_dhcp() {
    [[ -n ${CERANA_MGMT_MAC} ]] \
        && [[ -n ${MAC_TO_IFACE[${CERANA_MGMT_MAC}]} ]] \
        || return 1
    echo -e "[Link]\nMACAddress=${CERANA_MGMT_MAC}\n\n[Network]\nDHCP=yes" >/data/config/network/mgmt.network
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
                echo "\"server\": \"$server\"" >>$CONFIG
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

## Main
collect_addresses

if [[ -n $CERANA_CLUSTER_BOOTSTRAP ]]; then
    # We're bootstrapping a layer 2 cluster
    # 1. Which MAC address are we using
    config_mgmt_dhcp \
        || query_interface
    # 2. What IP range are we using?
    query_ip
    drop_consul_config bootstrap \
        || fail_exit $? "Cluster configuration failed"

elif [[ -n $CERANA_CLUSTER_IPS ]]; then
    # We're joining a layer 2 cluster
    # We should have been told which MAC to use for DHCP and which IPs to pass to consul
    config_mgmt_dhcp \
        || fail_exit $? "Cluster join expects a working mgmt interface to be specified"
    drop_consul_config join \
        || fail_exit $? "Cluster joining failed"

else
    # We're a standalone node
    # If we were given a MAC address to use on the kernel command line, use it
    # Otherwise, prompt for which device to use and ask about DHCP or static config
    config_mgmt_dhcp && exit 0

    #FIXME Only do DHCP for now. Static IP for standalone will come later.
    query_interface

fi

# remove the bootstrap flag for future boots
unset CERANA_CLUSTER_BOOTSTRAP

declare | grep ^CERANA >/data/config/cerana-bootcfg
config_mgmt_dhcp && exit 0

fail_exit 1 "Reached the end of the script without configuring any networking"
