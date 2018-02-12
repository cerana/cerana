{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaBundleHeartbeat;
  cfgdir = "/data/config/";
  cfgfile = "bundle-heartbeat.json";
  socketdir = "unix:///task-socket/node-coordinator/coordinator/";
  socket = "node-coord.sock";
  clusterDataURL = "unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock";
  utility = "${pkgs.cerana.bin}/bin/bundle-heartbeat";
in
{
  options.services.ceranaBundleHeartbeat.enable = mkEnableOption "ceranaBundleHeartbeat";

  config = mkIf cfg.enable {
    systemd.services.ceranaBundleHeartbeat = {
      description = "Cerana Bundle Heartbeat";
      wantedBy = [ "ceranaLayer2.target" ];
      after = [ "ceranaL2Coordinator.service"
                "ceranaClusterConfProvider.service"
                "ceranaMetricsProvider.service"
                "ceranaSystemdProvider.service"
                "ceranaZfsProvider.service"
                "systemd-networkd.service" ];
      serviceConfig = {
        Type = "simple";
        ExecStart = "${utility} -c ${cfgdir}${cfgfile}";
        Restart = "always";
        RestartSec = "3";
        TimeoutStopSec = "15";
      };
      preStart = ''
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "clusterDataURL": "${clusterDataURL}",' >> ${cfgdir}${cfgfile}
                echo '  "nodeDataURL": "${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "requestTimeout": "10s",' >> ${cfgdir}${cfgfile}
                echo '  "tickInterval": "5s",' >> ${cfgdir}${cfgfile}
                echo '  "tickRetryInterval": "5s"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
