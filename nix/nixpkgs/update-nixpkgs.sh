#!/usr/bin/env nix-shell
#!nix-shell -i bash -p nix-prefetch-git

nix-prefetch-git https://github.com/nixos/nixpkgs-channels.git \
	--no-deepClone \
	--rev refs/heads/nixos-17.09 > $(dirname $0)/nixpkgs.json
