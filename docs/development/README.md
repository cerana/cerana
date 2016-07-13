# Cerana Development Tools

## Setting up a development VM

* WARNING: The default boot option on the ISO will wipe your disks and automatically create a zpool on them.

* Create a Virtual Machine giving it:
  * Provide a minimum of 3GB of memory
  * Provide as many VCPUs as you'd like.
  * A disk (virtio for best performance) in this example it will be "/dev/vda"
  * A network interface (virtio for best performance) on a network with a DHCP server 
  * A virtual CD drive with the latest CeranaOS ISO image in it.
  * A virtual serial port to access the console (GRUB and the configuration prompts will show up here)
    * use COM1/ttyS0
* Set the VM to boot from the CD
* Boot the VM and connect to the serial port
* The default boot option will create a zpool on your disk and do DHCP on your singular network interface.
* The system will come up with only the Cerana Layer 1 services running.
* There is no root password, type "root" as your username and you will be logged in.

### Under SmartOS

* Example configuration (update the nic_tag as appropriate for your system):
```json

  "alias": "cerana-dev-node",
  "autoboot": "false",
  "brand": "kvm",
  "ram": 3072,
  "vcpus": 3,
  "boot": "order=dc",
  "disks": [
    {
      "path": "/cerana.iso",
      "media": "cdrom",
      "model": "ide",
      "boot": true
    },
    {
      "size": 2000,
      "compression": "lz4",
      "model": "virtio"
    }
  ],
  "nics": [
    {
      "nic_tag": "external",
      "model": "virtio",
      "ips": [ "dhcp" ],
      "primary": true
    }
  ]
}
```
* Create the VM
* Copy the ISO into `/zones/<uuid>/root/cerana.iso`
* `vmadm start <uuid>`
* `vmadm console <uuid>` to connect to the console

### Under Virtualbox (tested on OSX)

* Install `socat`
* New VM
  * Give it a Name
  * Type: Linux
  * Version: Linux 2.6 / 3.x /4.x (64-bit)
* Memory size: 3072 or greater
* Create a virtual hard disk now (whatever type you prefer, and dynamic or fixed doesn't matter from the CeranaOS perspective)
  * Give it at least a few GB, but CeranaOS itself currently has minimal disk needs, size it according to what you'll be doing with the VM.
* Open up the settings for the VM
  * Under System, give it more CPUs depending on planned workload
  * Under Storage, Put the CeranaOS ISO into the virtual CD drive
  * Under Network, for Adapter 1, under Advanced
    * Change the Adapter Type to "Paravirtualized Network (virtio-net)" for performance
    * To set up SSH forwarding, click the Port Forwarding button and in the new window:
      * On the top right click the plus button to add a rule and edit it
      * Name: `SSH`, Protocol: leave as `TCP`, Host Port: you choose, e.g. `2222`, Guest Port: `22`, Leave Host and Guest IP blank
  * To switch to bridged networking (optional):
    * Set "Attached to" to "Bridged Adapter"
    * Pick which physical NIC on the local machine you want to be bridged through (e.g. wifi vs ethernet)
  * Under Ports, Under Serial Ports, Port 1
    * Click the Enable box
    * Set the Port Mode to TCP
    * Make sure the "connect ot existing pipe/socket" box is NOT checked
    * Put a port (e.g. 12345) into the Path/Address box.
* Boot the VM
* Connect to the serial console with (fix the port number to what you used above) e.g. `socat STDIO,rawer TCP4:localhost:12345`

### Under Linux QEMU/KVM

FIXME

## Setting up a persistent local user account and persistent SSH host keys

* The development builds of CeranaOS include a running `sshd`. To have those survive reboots:
```
rsync -av /etc/ssh/ /data/config/ssh/
find /data/config/ssh/ -type l -delete
cat >/data/services/persist-ssh-keys.service <<"EOF"
[Unit]
Before=sshd.service
WantedBy=cerana.target
Description=Restore SSH Host Keys

[Service]

ExecStart=/run/current-system/sw/bin/rsync -av /data/config/ssh/ /etc/ssh/
Type=simple
EOF
ln -s ../persist-ssh-keys.service /data/services/cerana.target.wants/persist-ssh-keys.service
```
* Create a persistent `/home`
```
zfs create -o mountpoint=/home data/home
```
* Create a service to recreate your user at boot time (remember to adjust the username and password)
```
cat > /data/services/localusers.service <<"EOF"
[Unit]
Before=cerana.target
WantedBy=cerana.target
Description=Local User Setup

[Service]
Environment=USERNAME=myuser
Environment=PASSWORD=password123
Environment=UID=1000
ExecStartPre=/bin/sh -c "/run/current-system/sw/bin/groupadd -g $UID $USERNAME"
ExecStartPre=/bin/sh -c "/run/current-system/sw/bin/useradd -g $USERNAME -G wheel -u $UID -m -d /home/$USERNAME $USERNAME"
ExecStart=/bin/sh -c "/run/current-system/sw/bin/chpasswd <<<$USERNAME:$PASSWORD"
Type=simple
EOF
ln -s ../localusers.service /data/services/cerana.target.wants/localusers.service
systemctl daemon-reload
systemctl start localusers
```

## Rapid Updates for Develoment Machines

Starting at build 83 we are shipping a pair of tools for doing rapid updates of a CeranaOS machine from our S3 bucket of development builds.
You can run `cerana-update-dev-platform` to download the latest kernel and initrd files.
Once that completes, you can run `fastreboot` which will use kexec to rapidly reboot using them. The VM will see the same /proc/cmdline as it did on the previous boot.

## Building CeranaOS on CeranaOS

As root you can run `create-build-container` which will fetch a NixOS liveCD ISO and use the contents to set up a minimal build container, as well as fetching a copy of our cerana/nixpkgs repo.
When it's finished you can run `enter-build-container` which will use systemd-nspawn to set up the appropriate namespaces and drop you at a shell.
That shell behaves oddly, but running `exec bash -i -l` seems to clear up most of those quirks.
If you plan on using screen, running `export SHELL` gets things set to use bash as you would expect.

If all you want to do is run a build:
```bash
cd /nixpkgs
nix-build -A netboot nixos/release.nix
```

To make the build use more parallelism you may prefer our current preferred invocation:
```bash
time nix-build --cores 0 --max-jobs 3 -A netboot nixos/release.nix
```

Additionally, while use of the Nix package manager is a bit out of scope for this document, you can install some additional tools like so:
```
nix-env -f /nixpkgs/default.nix -i git go-1.6.2 go1.6-glide go1.6-shfmt ShellCheck-0.4.4 vim
```
