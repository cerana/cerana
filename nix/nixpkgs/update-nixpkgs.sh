#!/usr/bin/env nix-shell
#!nix-shell -i bash -p nix-prefetch-git

nix-prefetch-git https://github.com/nixos/nixpkgs.git \
	--no-deepClone \
	--rev refs/heads/nixos-unstable-small > $(dirname $0)/nixpkgs.json
