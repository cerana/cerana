#!/usr/bin/env bash

set -e

export MAKEFLAGS="-j$((2 * $(nproc) + 1))"

function fetch() {
    sudo apt-get update
    sudo apt-get -q -y -o "Dpkg::Options::=--force-confdef" -o "Dpkg::Options::=--force-confold" install \
        autoconf \
        bsdtar \
        dh-autoreconf \
        docker-engine \
        linux-headers-$(uname -r) \
        realpath \
        tree \
        uuid-dev \
        zlib1g-dev

    cd /tmp
    if [ ! -d "$ZFS_CACHE/spl-$VZFS" ]; then
        curl -L https://github.com/zfsonlinux/zfs/releases/download/zfs-$VZFS/spl-$VZFS.tar.gz | bsdtar -xf- -C $ZFS_CACHE
    fi
    if [ ! -d "$ZFS_CACHE/zfs-$CZFS" ]; then
        curl -L https://github.com/zfsonlinux/zfs/archive/$CZFS.tar.gz | bsdtar -xf- -C $ZFS_CACHE
    fi

    curl -L "https://releases.hashicorp.com/consul/$VCONSUL/consul_${VCONSUL}_linux_amd64.zip" | bsdtar -xf- -C$HOME/bin
    chmod +x $HOME/bin/consul

    curl -L "https://github.com/coreos/etcd/releases/download/v$VETCD/etcd-v$VETCD-linux-amd64.tar.gz" \
        | bsdtar -xf- -C$HOME/bin --strip-components=1 etcd-v$VETCD-linux-amd64/etcd

    curl -L "https://github.com/Masterminds/glide/releases/download/$VGLIDE/glide-$VGLIDE-linux-amd64.tar.gz" \
        | bsdtar -xf- -C$HOME/bin --strip-components=1 etcd-v$VETCD-linux-amd64/etcd

    go get github.com/alecthomas/gometalinter
    gometalinter --install --update
}

function install() {
    cd $ZFS_CACHE
    BUILT_FILE="$ZFS_CACHE/built-$VZFS-$CZFS"

    BUILT=0
    if [ -f "$BUILT_FILE" ] && (grep -xq "$(uname -rv)" "$BUILT_FILE"); then
        echo "Cache exists for spl and zfs, skipping build"
        BUILT=1
    else
        echo "No cache for spl and zfs, building"
    fi

    if [ $BUILT -eq 0 ]; then
        (
            cd spl-$VZFS
            ./configure --prefix=/usr
            make
        )
    fi

    (
        cd spl-$VZFS
        sudo make install
    )

    if [ $BUILT -eq 0 ]; then
        (
            cd zfs-$CZFS
            ./autogen.sh
            ./configure --prefix=/usr --with-spl=/usr/src/spl-$VZFS
            make
        )
    fi

    (
        cd zfs-$CZFS
        sudo make install
    )

    uname -rv >"$BUILT_FILE"
}

case $1 in
    fetch | install) ;;
    *)
        echo "unknown action: $1" >&2
        exit 1
        ;;
esac
$1
