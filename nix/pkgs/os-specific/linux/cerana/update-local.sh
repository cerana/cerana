#!/usr/bin/env nix-shell
#!nix-shell -i bash -p nix-prefetch-git git gawk

HERE="$(cd $(dirname $0); pwd)"
WORKTREE="$(cd ${HERE}; git worktree list --porcelain | awk '/worktree/{print $2}')"

if [[ -n ${1} ]]; then
    BRANCH="refs/heads/${1}"
else
    BRANCH="$(cd ${HERE}; git worktree list --porcelain | awk '/branch/{print $2}')"
fi

nix-prefetch-git "${WORKTREE}" --rev "${BRANCH}" > "${HERE}/cerana.json"
