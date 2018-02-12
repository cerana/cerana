{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaPlatformImport;
in
{
  options.services.ceranaPlatformImport.enable = mkEnableOption "Cerana Platform Import";

  config = mkIf cfg.enable {
    systemd.services.ceranaPlatformImport = {
      description = "Cerana Platform Import";
      path = [ pkgs.cerana-scripts pkgs.rsync pkgs.utillinux pkgs.coreutils pkgs.curl ];
      wantedBy = [ "multi-user.target" ];
      after = [ "ceranapool.service" "systemd-networkd.service" ];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = "${pkgs.cerana-scripts}/scripts/copy-platform.sh";
        TimeoutStartSec = "5min";
        RemainAfterExit = true;
      };
    };
  environment.systemPackages = [ pkgs.cerana-scripts ];
  };
}
