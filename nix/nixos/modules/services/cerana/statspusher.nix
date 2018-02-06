{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaStatsPusher;
  cfgdir = "/data/config/";
  cfgfile = "statspusher.json";
  socketdir = "unix:///task-socket/node-coordinator/coordinator/";
  socket = "node-coord.sock";
  clusterDataURL = "unix:///task-socket/l2-coordinator/coordinator/l2-coord.sock";
  datasetDir = "data/datasets";
  utility = "${pkgs.cerana.bin}/bin/statspusher";
in
{
  options.services.ceranaStatsPusher.enable = mkEnableOption "ceranaStatsPusher";

  config = mkIf cfg.enable {
    systemd.services.ceranaStatsPusher = {
      description = "Cerana Statistics Pusher";
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
                echo '  "bundleInterval": "5s",' >> ${cfgdir}${cfgfile}
                echo '  "datasetInterval": "5s",' >> ${cfgdir}${cfgfile}
                echo '  "datasetDir": "${datasetDir}",' >> ${cfgdir}${cfgfile}
                echo '  "nodeInterval": "5s",' >> ${cfgdir}${cfgfile}
                echo '  "nodeDataURL": "${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "clusterDataURL": "${clusterDataURL}",' >> ${cfgdir}${cfgfile}
                echo '  "requestTimeout": "10s"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
