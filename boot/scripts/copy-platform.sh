#!/bin/bash -x

# shellcheck disable=SC1091
source /tmp/cerana-bootcfg

# Don't do this in rescue mode
[[ -n $CERANA_RESCUE ]] && exit 0

verify_existing() {
    [[ -n "${CERANA_INITRD_HASH}" ]] || return 1 # no checksum to verify
    DESTINATION="/data/platform/${CERANA_INITRD_HASH}"
    IFS=- read -r ALG HASH <<<"${CERANA_INITRD_HASH}"
    type "${ALG}sum" || return 2                                # don't know how to verify this checksum
    "${ALG}sum" -c <<<"${HASH}  ${DESTINATION}/initrd" || return 3 # checksum failed
}

link_platform() {
    rm -f /data/platform/current \
        && ln -s "${DESTINATION#/data/platform/}" /data/platform/current
}

fetch_platform() {
    SERVER=${CERANA_BOOT_SERVER%/} # don't depend on a trailing slash
    mkdir -p "${DESTINATION}"
    until curl -o "${DESTINATION}/bzImage" "${SERVER}/kernel"; do
        sleep 1
    done
    curl -o "${DESTINATION}/initrd" "${SERVER}/initrd" || exit 1
}

if [[ -n "${CERANA_BOOT_SERVER}" ]]; then
    # We were told where to get boot media
    if [[ -n "${CERANA_INITRD_HASH}" ]]; then
        # We know the checksum of the initrd we booted with too
        verify_existing && link_platform && exit 0 # already got it!
        fetch_platform
        verify_existing \
            && link_platform
        exit $?
    else
        # If we didn't get a hash to use, use a default destination
        # and just trust curl
        DESTINATION=/data/platform/download
        fetch_platform
        link_platform
        exit $?
    fi
elif [[ -L /dev/disk/by-label/CERANA ]]; then
    # The CD/USB is in the drive
    DESTINATION=/data/platform/bootcd
    mkdir -p /mnt
    mount /dev/disk/by-label/CERANA /mnt
    mkdir -p "${DESTINATION}"
    rsync -a /mnt/ "${DESTINATION}/" \
        && link_platform
    ret=$?
    umount /mnt
    exit ${ret}
else
    # No boot server provided, no boot media provided.
    # Exit based on existence of media in the pool already
    [[ -f /data/platform/current/bzImage ]] \
        && [[ -f /data/platform/current/initrd ]]
    exit $?
fi
