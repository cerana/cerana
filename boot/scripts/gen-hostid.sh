#!/bin/bash
# Get the SMBIOS product_uuid and use it for the basis of hostid and machine-id
# Parts of this script is based on one from Fazle Arefin

uuid_file=/sys/class/dmi/id/product_uuid

if [ -f $uuid_file ]; then
    uuid=$(tr '[:upper:]' '[:lower:]' <$uuid_file)
else
    uuid=$(uuidgen)
fi

host_id=${uuid:0:8}
machine_id=$(tr -d - <<<$uuid)

a=${host_id:6:2}
b=${host_id:4:2}
c=${host_id:2:2}
d=${host_id:0:2}

echo -ne \\x$a\\x$b\\x$c\\x$d >/etc/hostid \
    && echo "Setting hostid to $host_id"

echo $machine_id >/etc/machine-id \
    && echo "Setting machine-id to $machine_id"

echo $uuid >/etc/machine_id \
    && echo "Setting machine_id to $uuid"

echo $uuid >/etc/hostname \
    && echo "Setting hostname to $uuid"

exit 0
