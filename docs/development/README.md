Cerana Development Tools
=================

Setting up a development VM
--------------------

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

