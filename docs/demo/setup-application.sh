#/bin/bash

set -o xtrace

l2-request () { coordinator-cli -c http://172.16.10.2:8085 -r :4080 "$@"; }
import-dataset () { l2-request -t import-dataset -s -a redundancy=${1} -a readOnly=true ; }
update-service () { l2-request -t update-service -j ; }
update-bundle () { l2-request -t update-bundle -j ; }

service () {
cat <<EOF
{
  "service": {
    "dataset": "${1}",
    "cmd": [
      ${2}
    ],
    "env": {
      "PORT": "${3}"
    }
  }
}
EOF
}

bundle () {
cat <<EOF
{
  "bundle": {
    "services": {
      "${1}": { "id": "${1}" }
    }
  }
}
EOF
}

DATASET1=$(import-dataset 1 < backend-in-a-zfs-stream.zfs | tee /dev/stderr | jq -r .dataset.id)
[[ -n ${DATASET1} ]] && SERVICE1=$(service ${DATASET1} '"/backend"' 10000 | tee /dev/stderr | update-service | tee /dev/stderr | jq -r .service.id)
[[ -n ${SERVICE1} ]] && bundle ${SERVICE1} | tee /dev/stderr | update-bundle

DATASET2=$(import-dataset 1 < haproxy-in-a-zfs-stream.zfs | tee /dev/stderr | jq -r .dataset.id)
[[ -n ${DATASET2} ]] && SERVICE2=$(service ${DATASET2} '"/haproxy.py"' 8001 | tee /dev/stderr | update-service | tee /dev/stderr | jq -r .service.id)
[[ -n ${SERVICE2} ]] && bundle ${SERVICE2} | tee /dev/stderr | update-bundle
