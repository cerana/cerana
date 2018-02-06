{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaHealthProvider;
  name = "health-provider";
  cfgdir = "/data/config/";
  cfgfile = "health-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/health-provider";
in
{
  options.services.ceranaHealthProvider.enable = mkEnableOption "ceranaHealthProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaHealthProvider = {
      description = "Cerana Health Provider";
      wantedBy = [ "multi-user.target" ];
      wants = [ "ceranaSystemdProvider.service" ];
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
