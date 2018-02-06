#!/bin/sh -e

cd "$(dirname "$0")" || exit

overlay_dir=$HOME/.config/nixpkgs/overlays
name=cerana

echo Installing $name as an overlay

set -x
mkdir -p "$overlay_dir"
ln -s "$PWD" "$overlay_dir/$name"
