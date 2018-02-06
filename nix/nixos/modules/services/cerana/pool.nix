{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranapool;
in
{
  options.services.ceranapool.enable = mkEnableOption "Cerana Pool Configuration";

  config = mkIf cfg.enable {
    systemd.services.ceranapool = {
      description = "Cerana Pool Configuration";
      path = [ pkgs.cerana-scripts pkgs.gawk pkgs.utillinux pkgs.gnugrep
        pkgs.grub2 pkgs.gptfdisk pkgs.systemd pkgs.zfs pkgs.bash pkgs.coreutils];
      requiredBy = [ "network-pre.target" "multi-user.target" "sysinit.target"
        "local-fs-pre.target" "zfs-import.target" "systemd-journald.service"
        "basic.target" "ntpd.service" "sshd.service" "time-sync.service"
        "getty.target" "ip-up.target" "ceranaMoveLogs.service"];
      before = [ "sshd.service" "ntpd.service" "local-fs.target"
        "ceranaMoveLogs.service"];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = "${pkgs.cerana-scripts}/scripts/init-zpools.sh";
        TimeoutStartSec = "5min";
        StandardInput = "tty";
        TTYPath = "/dev/console";
        TTYReset = "yes";
        TTYVHangup = "yes";
        RemainAfterExit = true;
      };
    };
  environment.systemPackages = [ pkgs.cerana-scripts ];
  };
}
