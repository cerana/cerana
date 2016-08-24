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

gateway_string() {
    [[ -n ${CERANA_MGMT_GW} ]] \
        && echo "\"gateway\":\"${CERANA_MGMT_GW}\","
}

l2-request() { coordinator-cli -c unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock -r :4080 "$@"; }

until l2-request -t kv-keys -a key=/ &>/dev/null; do
    sleep 1
done

until l2-request -t kv-get -a key=/dhcp &>/dev/null; do
    l2-request -t set-dhcp-config -j <<EOF
{
  "duration": "12h",
  $(gateway_string)
  "net": "${CERANA_MGMT_IP}"
}
EOF
    sleep 1
done

LEASE_JSON="{\"mac\":\"${CERANA_MGMT_MAC}\", \"ip\":\"${CERANA_MGMT_IP%/*}\"}"
until l2-request -t dhcp-ack-lease -j <<<"${LEASE_JSON}"; do
    sleep 1
done

unset CERANA_CLUSTER_BOOTSTRAP
declare | grep ^CERANA >/data/config/cerana-bootcfg
