# This module defines the software packages included in the "minimal"
# cerana image
{ config, lib, pkgs, ... }:

{
  # Include some utilities that are useful for installing or repairing
  # the system.
  environment.systemPackages = [
    pkgs.gptfdisk

    # Hardware-related tools.
    pkgs.sdparm
    pkgs.hdparm
    pkgs.smartmontools # for diagnosing hard disks
    pkgs.pciutils
    pkgs.usbutils

    pkgs.grub2
    pkgs.libselinux
    pkgs.iptables
    pkgs.strace
    pkgs.lshw
    pkgs.cerana-scripts
    pkgs.dhcpcd
    pkgs.gptfdisk

    pkgs.git
    pkgs.vim
    pkgs.patchelf
    pkgs.screen
    pkgs.jq
  ];


  # Use more recent kernel
  boot.kernelPackages = pkgs.linuxPackages_latest;

  # Include support for various filesystems.
  boot.supportedFilesystems = [ "zfs" ];
  boot.zfs.forceImportAll = false;

  # Configure host id for ZFS to work
  networking.hostId = lib.mkDefault "8425e349";

  security.apparmor.enable = false;

  services.cerananet.enable = true;
  services.ceranaPlatformImport.enable = true;
  services.ceranapool.enable = true;
  services.ceranaMoveLogs.enable = true;
  targets.cerana.enable = true;

  nix.nrBuildUsers = 0;
  systemd.network.enable = true;
  networking.useDHCP = false;

  # don't use NixOS firewall
  networking.firewall.enable = false;

  # For development puroposes only
  services.sshd.enable = true;
}
