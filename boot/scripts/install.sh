#!/bin/bash

## This executes due to the install flag being passed on the commands line.
##
## It asks basic questions from the user regarding IP addresses, consul cluster
## IP addresses and so forth, and adds the inputted values to the boot configuration
## file.
##

CERANA_BOOTCFG=/tmp/cerana-bootcfg

CERANA_IFACE=""
CERANA_IP=""
CERANA_NETMASK=""
CERANA_GW=""
CERANA_CLUSTER_IPS=""

declare -a IFACES

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
	pushd .
	cd /sys/class/net
	IFACES=(e*)
	popd
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

## Main

query_interface
query_ip
query_gateway

echo "CERANA_IFACE=$CERANA_IFACE" >> $CERANA_BOOTCFG
echo "CERANA_IP=$CERANA_IP/$CERANA_NETMASK" >> $CERANA_BOOTCFG
echo "CERANA_GW=$CERANA_GW" >> $CERANA_BOOTCFG
