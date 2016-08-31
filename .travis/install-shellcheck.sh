#!/usr/bin/env bash

set -o errexit
set -o xtrace

SHELLCHECK_VERSION="${1}"
[[ -n "${SHELLCHECK_VERSION}" ]]

check_and_install() {
    ~/.cabal/bin/shellcheck --version | grep -q "${SHELLCHECK_VERSION}"
    sudo install -D -m 755 -t /usr/local/bin ~/.cabal/bin/shellcheck
    exit 0
}

if [[ -x ~/.cabal/bin/shellcheck ]]; then
    # If we have a cached version sitting around
    check_and_install
else
    # Build it!
    sudo apt-get update
    sudo apt-get install cabal-install
    curl -L "https://github.com/koalaman/shellcheck/archive/v${SHELLCHECK_VERSION}.tar.gz" | tar xz
    cd "shellcheck-${SHELLCHECK_VERSION}"
    cabal update
    cabal install
    [[ -x ~/.cabal/bin/shellcheck ]]
    check_and_install
fi

# Something went wrong.
exit 1
