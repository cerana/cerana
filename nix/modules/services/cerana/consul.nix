{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaConsul;
in
{
  options.services.ceranaConsul.enable = mkEnableOption "ceranaConsul";

  config = mkIf cfg.enable {
    systemd.services.ceranaConsul = {
      description = "Cerana Consul";
      wantedBy = [ "ceranaLayer2.target" ];
      after = [ "ceranaNodeCoordinator.service" ];
      serviceConfig = {
        Type = "simple";
        KillMode = "process";
        KillSignal = "SIGINT";
        ExecStart = "${pkgs.consul.bin}/bin/consul agent -config-file=/data/config/consul.json";
        Restart = "always";
        RestartSec = "3";
      };
    };
  };
}
