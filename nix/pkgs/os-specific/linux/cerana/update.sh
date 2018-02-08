#!/usr/bin/env nix-shell
#!nix-shell -i bash -p nix-prefetch-git

HERE=$(cd $(dirname $0); pwd)

if [[ -n ${1} ]]; then
    HEAD=${1}
else
    HEAD=master
fi

nix-prefetch-git https://github.com/cerana/cerana.git \
                 --rev refs/heads/${HEAD} > ${HERE}/cerana.json
