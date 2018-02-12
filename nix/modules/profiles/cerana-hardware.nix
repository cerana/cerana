# This module enables all hardware supported by NixOS: i.e., all
# firmware is included, and all devices from which one may boot are
# enabled in the initrd.  Its primary use is in the NixOS installation
# CDs.

{ config, pkgs, ... }:

{

  # The initrd has to contain any module that might be necessary for
  # supporting the most important parts of HW like drives.
  boot.initrd.availableKernelModules =
    [ # SATA/PATA support.
      "ahci"

      "ata_piix"

      "sata_inic162x" "sata_nv" "sata_promise" "sata_qstor"
      "sata_sil" "sata_sil24" "sata_sis" "sata_svw" "sata_sx4"
      "sata_uli" "sata_via" "sata_vsc"

      # SCSI support (incomplete).
      "3w-9xxx" "3w-xxxx" "aic79xx" "aic7xxx" "arcmsr"

      # USB support, especially for booting from USB CD-ROM
      # drives.
      "usb_storage"

      # Virtio (QEMU, KVM etc.) support.
      "virtio_net" "virtio_pci" "virtio_blk" "virtio_scsi" "virtio_balloon" "virtio_console"

      # Keyboards
      "usbhid" "hid_apple" "hid_logitech_dj" "hid_lenovo_tpkbd" "hid_roccat"
    ];

  # Include lots of firmware.
  hardware.enableAllFirmware = false;

}
