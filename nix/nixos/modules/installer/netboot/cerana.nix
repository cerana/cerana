# This module creates netboot media containing the given NixOS
# configuration.

{ config, lib, pkgs, ... }:

with lib;

{

  config = {

    boot.initrd.compressor = "${pkgs.pigz}/bin/pigz -R -b 1024 ";
    boot.loader.grub.version = 2;

    # Don't build the GRUB menu builder script, since we don't need it
    # here and it causes a cyclic dependency.
    boot.loader.grub.enable = false;

    # !!! Hack - attributes expected by other modules.
    system.boot.loader.kernelFile = "bzImage";

    fileSystems."/" =
      { fsType = "tmpfs";
        options = [ "mode=0755" ];
      };

    # Create the initrd
    system.build.netbootRamdisk = pkgs.makeInitrd {
      inherit (config.boot.initrd) compressor;

      contents =
        [ { object = config.system.build.toplevel + "/init";
            symlink = "/init";
          }
        ];
    };

    # Create the docker image
    system.build.netbootDockerImage = pkgs.dockerTools.buildImage {
      name = "netboot";
      contents = config.system.build.toplevel;
    };

    system.build.netbootIpxeScript = pkgs.writeTextDir "netboot.ipxe" "#!ipxe\nkernel bzImage console=ttyS0 vga=794 cerana.mgmt_mac=\${mac} ${toString config.boot.kernelParams}\ninitrd initrd\nboot";

    system.build.ceranaGrubConfig = pkgs.writeTextDir "menu.lst"
''
default 0
timeout 10
min_mem64 1024
serial
terminal serial

title CeranaOS Standalone Automatic ZFS
   kernel /bzImage ${toString config.boot.kernelParams} cerana.zfs_config=auto console=ttyS0
   module /initrd

title CeranaOS Standalone Pool Prompt
   kernel /bzImage ${toString config.boot.kernelParams} console=ttyS0
   module /initrd

title CeranaOS Rescue Mode
   kernel /bzImage ${toString config.boot.kernelParams} cerana.rescue console=ttyS0
   module /initrd

title CeranaOS Cluster Bootstrap (Automatic ZFS, 172.16.10.2/16)
   kernel /bzImage ${toString config.boot.kernelParams} cerana.cluster_bootstrap cerana.zfs_config=auto cerana.mgmt_ip=172.16.10.2/16 console=ttyS0
   module /initrd

title CeranaOS Cluster Join (iPXE)
   kernel /ipxe.lkrn

title Boot from first HDD
   rootnoverify (hd0)
   chainloader +1
'';

    system.build.ceranaGrub2Config = pkgs.writeTextDir "grub.cfg"
''
serial --unit=0 --speed=115200
terminal_input serial console
terminal_output serial console

set color_normal=white/black
set color_highlight=black/white

set default=0
set timeout=10

search --set=root --label CERANA

menuentry "CeranaOS Standalone Automatic ZFS" {
   linux /bzImage ${toString config.boot.kernelParams} cerana.zfs_config=auto console=ttyS0
   initrd /initrd
}

menuentry "CeranaOS Standalone Pool Prompt" {
   linux /bzImage ${toString config.boot.kernelParams} console=ttyS0
   initrd /initrd
}

menuentry "CeranaOS Rescue Mode" {
   linux /bzImage ${toString config.boot.kernelParams} cerana.rescue console=ttyS0
   initrd /initrd
}

menuentry "CeranaOS Cluster Bootstrap (Automatic ZFS, 172.16.10.2/16)" {
   linux /bzImage ${toString config.boot.kernelParams} cerana.cluster_bootstrap cerana.zfs_config=auto cerana.mgmt_ip=172.16.10.2/16 console=ttyS0
   initrd /initrd
}

menuentry "CeranaOS Cluster Join (iPXE)" {
   linux16 /ipxe.lkrn
}

menuentry "Boot from first HDD" {
   set root=(hd0)
   chainloader +1
}
'';

    boot.loader.timeout = 10;

    boot.postBootCommands =
      ''
        ${pkgs.procps}/bin/sysctl -w net.core.rmem_max=8388608
        ${pkgs.procps}/bin/sysctl -w net.core.wmem_max=8388608
        ${pkgs.coreutils}/bin/rm /etc/hostid
        ${pkgs.coreutils}/bin/mkdir -p /task-socket/node-coordinator/
        ${pkgs.cerana-scripts}/scripts/parse-cmdline.sh
        ${pkgs.cerana-scripts}/scripts/gen-hostid.sh
      '';

  };

}
