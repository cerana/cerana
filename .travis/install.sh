#!/usr/bin/env bash

set -e

export MAKEFLAGS="-j$((2 * $(nproc) + 1))"

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
	zlib1g-dev \
	;

cd /tmp

curl -L https://github.com/zfsonlinux/zfs/releases/download/zfs-$VZFS/spl-$VZFS.tar.gz | bsdtar -xf-
pushd spl-$VZFS
./configure --prefix=/usr
make
sudo make install
popd

curl -L https://github.com/zfsonlinux/zfs/archive/$CZFS.tar.gz | bsdtar -xf-
pushd zfs-$CZFS
./autogen.sh
./configure --prefix=/usr --with-spl=/usr/src/spl-$VZFS
make
sudo make install
popd

curl -L "https://releases.hashicorp.com/consul/$VCONSUL/consul_${VCONSUL}_linux_amd64.zip" | bsdtar -xf- -C$HOME/bin
chmod +x $HOME/bin/consul

curl -L "https://github.com/coreos/etcd/releases/download/v$VETCD/etcd-v$VETCD-linux-amd64.tar.gz" | \
	bsdtar -xf- -C$HOME/bin --strip-components=1 etcd-v$VETCD-linux-amd64/etcd
