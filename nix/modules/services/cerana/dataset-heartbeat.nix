{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaDatasetHeartbeat;
  cfgdir = "/data/config/";
  cfgfile = "dataset-heartbeat.json";
  socketdir = "unix:///task-socket/node-coordinator/coordinator/";
  socket = "node-coord.sock";
  clusterDataURL = "unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock";
  datasetDir = "data/datasets";
  utility = "${pkgs.cerana.bin}/bin/dataset-heartbeat";
in
{
  options.services.ceranaDatasetHeartbeat.enable = mkEnableOption "ceranaDatasetHeartbeat";

  config = mkIf cfg.enable {
    systemd.services.ceranaDatasetHeartbeat = {
      description = "Cerana Dataset Heartbeat";
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
                echo '  "datasetPrefix": "${datasetDir}",' >> ${cfgdir}${cfgfile}
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
