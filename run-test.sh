#!/usr/bin/env bash

set -e

# args:
# $1: relative path to .test file

dir=$(dirname "$1")
name=$(basename "$1")
exec 2> >(tee "$dir/test.out")

which consul &>/dev/null

cid=$(docker run -dti \
    --cap-add SYS_ADMIN \
    --device /dev/zfs:/dev/zfs \
    --env KV="${KV:-consul}" \
    --name "$name" \
    --volume "$(mktemp -d):/tmp" \
    --volume "$PWD:/mistify:ro" \
    --volume /sys/fs/cgroup:/sys/fs/cgroup:ro \
    mistifyio/mistify-os:zfs-stable-api
)

[[ -n $cid ]]
sleep .25

docker cp "$(which consul)" "$cid:/usr/bin/"
docker cp "$(which etcd)" "$cid:/usr/bin/"
docker exec -i "$cid" sh -c "cd '/mistify/$dir'; echo '### TEST  $name'; './$name' -test.v" >&2
ret=$?

docker kill "$cid" >/dev/null || :
docker rm -v "$cid" >/dev/null || :
exit $ret
