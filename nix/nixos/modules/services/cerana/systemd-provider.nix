{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaSystemdProvider;
  name = "systemd-provider";
  cfgdir = "/data/config/";
  cfgfile = "systemd-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/systemd-provider";
in
{
  options.services.ceranaSystemdProvider.enable = mkEnableOption "ceranaSystemdProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaSystemdProvider = {
      description = "Cerana Systemd Provider";
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
                echo '  "coordinator_url": "unix://${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "unit_file_dir": "/data/services"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
