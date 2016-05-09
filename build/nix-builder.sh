#!/bin/bash
# This sets the path to point to the nix tools.
source ~/.nix-profile/etc/profile.d/nix.sh

export LC_ALL=C
time nix-build $*
if [ -h result/etc ]; then echo Error: Build resulted /etc as symlink && exit 1; fi
nix-store -q result --graph | sed 's/#ff0000/#ffffff/' | dot -Nstyle=bold -Tpng > dependencies.png
nix-store -q result --tree > dependencies.txt
