#!/usr/bin/env bash

set -e

# args:
# $1: relative path to .test file

name=$(basename "$1")
exec 2> >(tee "$(dirname "$1")/test.out")

which consul &>/dev/null

cid=$(docker run -dti \
    --cap-add SYS_ADMIN \
    --device /dev/zfs:/dev/zfs \
    --env TMPDIR="$(mktemp -d)" \
    --env KV="${KV:-consul}" \
    --name "$name" \
    --volume "$PWD:/mistify:ro" \
    --volume /sys/fs/cgroup:/sys/fs/cgroup:ro \
    --volume /tmp:/tmp \
    mistifyio/mistify-os:zfs-stable-api
)

[[ -n $cid ]]
sleep .25

docker cp "$(which consul)" "$cid:/usr/bin/"
docker cp "$(which etcd)" "$cid:/usr/bin/"
docker exec -i "$cid" sh -c "echo '### TEST  $name'; '/mistify/$1' -test.v" >&2
ret=$?

docker kill "$cid" >/dev/null || :
docker rm -v "$cid" >/dev/null || :
exit $ret
