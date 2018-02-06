{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaBootstrap;
in
{
  options.services.ceranaBootstrap.enable = mkEnableOption "Cerana Cluster Bootstrapping";

  config = mkIf cfg.enable {
    systemd.services.ceranaBootstrap = {
      description = "Cerana Cluster Bootstrapping";
      path = [ pkgs.cerana-scripts pkgs.cerana.bin pkgs.coreutils ];
      requires = [ "ceranaL2Coordinator.service" "ceranaKvProvider.service" "ceranaDhcpProvider.service" "ceranaClusterConfProvider.service" ];
      after = [ "ceranaL2Coordinator.service" "ceranaKvProvider.service" "ceranaDhcpProvider.service" "ceranaClusterConfProvider.service" ];
      requiredBy = [ "ceranaBootserver.service" ];
      before = [ "ceranaBootserver.service" ];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = "${pkgs.cerana-scripts}/scripts/bootstrap-clusterconfig.sh";
        TimeoutStartSec = "2min";
        RemainAfterExit = true;
      };
    };
  environment.systemPackages = [ pkgs.cerana-scripts ];
  };
}
