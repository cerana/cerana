#!/usr/bin/env nix-shell
#!nix-shell -i bash --pure -I nixpkgs=./ -p grub2 xorriso bash multipath-tools coreutils utillinux ipxe

IPXE=$(echo $buildInputs | sed 's| |\n|g' | grep ipxe)

grub-mkrescue \
    -o result-iso \
    -V CERANA \
    -- \
    -follow on \
    -pathspecs on \
    -add boot/grub/grub.cfg=result/grub.cfg \
    bzImage=result/bzImage \
    initrd=result/initrd \
    ipxe.lkrn=${IPXE}/ipxe.lkrn
