#!/bin/bash -x

set -o errexit
export GIT_MERGE_AUTOEDIT=no
git checkout demo
git fetch --all
git reset --hard origin/master
for branch in demo_outline daisy_services 349_tick daisy kv-is-leader; do
    git merge --no-ff origin/$branch
done
git push origin demo --force
git rev-parse origin/demo
