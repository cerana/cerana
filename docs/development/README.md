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

### SmartOS

* Example configuration (update the nic_tag as appropriate for your system):
```json
{
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

### Virtualbox

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
  * Under Network, for Adapter 1, under Advanced, change the Adapter Type to "Paravirtualized Network (virtio-net)" for performance
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

### Linux
