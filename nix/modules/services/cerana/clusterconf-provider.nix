{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaClusterConfProvider;
  name = "clusterconfig-provider";
  cfgdir = "/data/config/";
  cfgfile = "clusterconfig-provider.json";
  socketdir = "/task-socket/l2-coordinator/";
  socket = "coordinator/l2-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/clusterconfig-provider";
in
{
  options.services.ceranaClusterConfProvider.enable = mkEnableOption "ceranaClusterConfProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaClusterConfProvider = {
      description = "Cerana Cluster Configuration Provider";
      wantedBy = [ "ceranaLayer2.target" ];
      wants = [ "ceranaKvProvider.service" ];
      after = [ "ceranaL2Coordinator.service" ];
      serviceConfig = {
        Type = "simple";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
        Restart = "always";
        RestartSec = "3";
      };
      preStart = ''
        find ${socketdir} -iname \*${name}.sock -delete
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
