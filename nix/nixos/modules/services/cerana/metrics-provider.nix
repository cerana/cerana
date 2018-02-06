{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaMetricsProvider;
  name = "metrics-provider";
  cfgdir = "/data/config/";
  cfgfile = "metrics-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/metrics-provider";
in
{
  options.services.ceranaMetricsProvider.enable = mkEnableOption "ceranaMetricsProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaMetricsProvider = {
      description = "Cerana Metrics Provider";
      path = [ pkgs.lshw ];
      wantedBy = [ "multi-user.target" ];
      wants = [ "ceranaNodeCoordinator.service" ];
      after = [ "ceranaNodeCoordinator.service" ];
      serviceConfig = {
        Type = "simple";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
        Restart = "always";
        RestartSec = "3";
      };
      preStart = ''
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "service_name": "${name}",' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}",' >> ${cfgdir}${cfgfile}
                echo '  "coordinator_url": "unix://${socketdir}${socket}"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
