#!/usr/bin/env nix-shell
#!nix-shell -i bash --pure -p curl gnugrep jq gnused nix

set -o xtrace
set -o errexit

HERE=$(cd $(dirname $0); pwd)

if [[ -n ${1} ]]; then
    HEAD=heads/${1}
else
    HEAD=heads/master
fi
# allow sepcifying e.g. tags/demo-0.0.1
HEAD=${HEAD/heads\/tags/tags}

REV=$(curl --cacert /etc/ssl/certs/ca-certificates.crt https://api.github.com/repos/cerana/cerana/git/refs/${HEAD} | jq -r .object.sha)

sed -e "/rev/{s|\"[^\"]*\"|\"${REV}\"|}" -i ${HERE}/default.nix
sed -e "/sha256/{s|\"..........|\"1111111111|}" -i ${HERE}/default.nix

HASH=$(nix-build -A cerana-scripts ${HERE}/../../../../default.nix 2>&1 | grep hash | sed 's|.* hash ‘||;s|’ .*||')

[[ -n ${HASH} ]] \
    && sed -e "/sha256/{s|\"[^\"]*\"|\"${HASH}\"|}" -i ${HERE}/default.nix
