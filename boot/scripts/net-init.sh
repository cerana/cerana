#!/bin/bash

source /tmp/cerana-bootcfg

# don't configure networking in rescue mode
[[ -n $CERANA_RESCUE ]] && exit 0


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
function is_valid_ip()
{
	local ip=$1
	local stat=1

	if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
		OIFS=$IFS
		IFS='.'
		ip=($ip)
		IFS=$OIFS
		[[ ${ip[0]} -le 255 && ${ip[1]} -le 255 && \
		${ip[2]} -le 255 && ${ip[3]} -le 255 ]]
		stat=$?
	fi
	return $stat
}

## is_valid_prefix <cidr prefix>
##
## Takes a CIDR prefix and makes sure that it is greater than 7 and less than 33
function is_valid_prefix()
{
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
	declare -a IFACES
	pushd .
	cd /sys/class/net
	IFACES=(e*)
	popd
}

# Create an associative array of interface and MAC addresses
function collect_addresses() {
    local name address nics
    declare -A nics
    for device in /sys/class/net/e*
    do
        name=${device#/sys/class/net/}
        read address < ${device}/address
        nics[$address]=$name
    done
    [[ -n ${CERANA_MGMT_MAC} ]] && CERANA_MGMT_IFACE=${nics[$CERANA_MGMT_MAC]}
}

function query_interface() {
	local answer

	iface_list
	echo
	echo "Detected interfaces:"
	echo
	for iface in ${IFACES[@]}; do
		printf "\t* %s\n" "$iface"
	done
	echo
	echo "Please select the main interface from the list above."
	answer=""
	echo -n "Selection: "
	read answer

	while [ ! $answer ] || ! in_array IFACES "$answer"; do
		echo "You must select from the listed interfaces!"
		echo -n "Selection: "
		read answer
	done

	CERANA_IFACE=$answer
}

function query_ip() {
	local answer
	local ip
	local prefix

	echo
	echo "Please specify the IP address and netmask of this node."
	echo
	echo -n "> "
	read answer

	ip=$(awk -F/ '{print $1}' <<< $answer)
	prefix=$(awk -F/ '{print $2}' <<< $answer)

	while [ ! $answer ] || \
	    [ ! $prefix ] || \
	    ! is_valid_ip $ip || \
	    ! is_valid_prefix $prefix; do
		echo "You must provide a valid IP address and prefix!"
		echo -n "> "
		read answer

		ip=$(awk -F/ '{print $1}' <<< $answer)
		prefix=$(awk -F/ '{print $2}' <<< $answer)
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
	[[ -n ${CERANA_MGMT_MAC} ]] || return 1
	echo -e "[Link]\nMACAddress=${CERANA_MGMT_MAC}\n\n[Network]\nDHCP=yes" > /data/config/networks/mgmt.network
}


function drop_consul_config() {
	echo "Configuring Consul not implemented yet, yell at Nahum"
	return 1
}


## Main
if [[ -n $CERANA_CLUSTER_BOOTSTRAP ]]; then
	# We're bootstrapping a layer 2 cluster
	# 1. Which MAC address are we using
	# 2. What IP range are we using?
	echo "Cluster bootstrap not implemented yet. Yell at Nahum"
	exit 1

elif [[ -n $CERANA_CLUSTER_IPS ]]; then
	# We're joining a layer 2 cluster
	# We should have been told which MAC to use for DHCP and which IPs to pass to consul
	config_mgmt_dhcp && \
	drop_consul_config
	exit $?
else
	# We're a standalone node
	# If we were given a MAC address to use on the kernel command line, use it
	# Otherwise, prompt for which device to use and ask about DHCP or static config
	config_mgmt_dhcp && exit 0

	#FIXME
	query_interface
	query_ip
	query_gateway

	export CERANA_IFACE CERANA_IP CERANA_GW
fi


export | grep CERANA > /data/config/cerana-bootcfg

# We don't need the following if we can get run before it starts
# systemctl reload systemd-networkd.service

# Still here???
echo "Yell at Nahum"
exit 1

