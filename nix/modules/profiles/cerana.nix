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
    pkgs.qemu_kvm
    pkgs.strace
    pkgs.gdb
    pkgs.lshw
    pkgs.consul.bin
    pkgs.cerana.bin
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

  services.ceranaBootserver.enable = true;
  services.ceranaBootstrap.enable = true;
  services.ceranaBundleHeartbeat.enable = true;
  services.ceranaClusterConfProvider.enable = true;
  services.ceranaConsul.enable = true;
  services.ceranaDatasetHeartbeat.enable = true;
  services.ceranaDataTradeProvider.enable = true;
  services.ceranaDhcpProvider.enable = true;
  services.ceranaHealthProvider.enable = true;
  services.ceranaKvProvider.enable = true;
  services.ceranaL2Coordinator.enable = true;
  services.ceranaMetricsProvider.enable = true;
  services.ceranaNamespaceProvider.enable = true;
  services.cerananet.enable = true;
  services.ceranaNodeCoordinator.enable = true;
  services.ceranaNodeHeartbeat.enable = true;
  services.ceranaPlatformImport.enable = true;
  services.ceranapool.enable = true;
  services.ceranaMoveLogs.enable = true;
  services.ceranaServiceProvider.enable = true;
  services.ceranaSystemdProvider.enable = true;
  services.ceranaZfsProvider.enable = true;
  targets.cerana.enable = true;
  targets.ceranaLayer2.enable = true;

  nix.nrBuildUsers = 0;
  systemd.network.enable = true;
  networking.useDHCP = false;

  # don't use NixOS firewall
  networking.firewall.enable = false;

  # For development puroposes only
  services.sshd.enable = true;
}
