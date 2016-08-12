#!/bin/bash -x

# shellcheck disable=SC1091
source /tmp/cerana-bootcfg

function exit_code_with_message() {
    local exit_code
    exit_code=${1}
    shift
    echo "${*}" >&2
    exit "${exit_code}"
}

[[ -n ${CERANA_RESCUE} ]] && exit_code_with_messsage 0 "Rescue Mode. Exiting."

[[ -z ${CERANA_CLUSTER_BOOTSTRAP} ]] && exit_code_with_message 0 "Not in bootstrap mode. Exiting."

[[ -z ${CERANA_MGMT_MAC} ]] && fail_exit 1 "Couldn't find MAC address for bootstrapping the cluster config. Failure."

[[ -z ${CERANA_MGMT_IP} ]] && fail_exit 1 "Couldn't find IP configuration for bootstrapping the cluster config. Failure."

LOCAL_IP=${CERANA_MGMT_IP%/*}
PREFIX_BITS=${CERANA_MGMT_IP#*/}

# FIXME UGLY HACK
case ${PREFIX_BITS} in
    8)
        MASK="/wAAAA=="
        NETMASK="255.0.0.0"
        ;;
    16)
        MASK="//8AAA=="
        NETMASK="255.255.0.0"
        ;;
    24)
        MASK="////AA=="
        NETMASK="255.255.255.0"
        ;;
    *)
        exit_code_with_message 2 "We only know how to handle /8, /16, /24 netmasks... sorry."
        ;;
esac

address_to_number() {
    local IFS=. ipStr
    ipStr=($1)
    echo $(($(($(($(($(($((ipStr[0] * 256)) + ipStr[1])) * 256)) + ipStr[2])) * 256)) + ipStr[3]))
}

number_to_address() {
    echo "$(($1 / 16777216)).$(($(($1 % 16777216)) / 65536)).$(($(($1 % 65536)) / 256)).$(($1 % 256))"
}

NET_ADDRESS=$(number_to_address $(($(address_to_number "${LOCAL_IP}") & $(address_to_number "${NETMASK}"))))

gateway_string() {
    [[ -n ${CERANA_MGMT_GW} ]] \
        && echo "\"gateway\":\"${CERANA_MGMT_GW}\","
}

l2-request() { coordinator-cli -c unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock -r :4080 "$@"; }

until l2-request -t kv-keys -a key=/ &>/dev/null; do
    sleep 1
done

l2-request -t set-dhcp -s <<EOF
{
  "duration": 86400000000000,
  $(gateway_string)
  "net": {
    "IP": "${NET_ADDRESS}",
    "Mask": "${MASK}"
  }
}
EOF

LEASE_JSON="{\"mac\":\"${CERANA_MGMT_MAC}\", \"ip\":\"${LOCAL_IP}\"}"
l2-request -t dhcp-offer-lease -s <<<"${LEASE_JSON}" \
    && l2-request -t dhcp-ack-lease -s <<<"${LEASE_JSON}"
