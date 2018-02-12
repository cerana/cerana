{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaNodeCoordinator;
  name = "node-coord";
  cfgdir = "/data/config/";
  cfgfile = "coordinator.json";
  socketdir = "/task-socket/node-coordinator/";
  daemon = "${pkgs.cerana.bin}/bin/coordinator";
in
{
  options.services.ceranaNodeCoordinator.enable = mkEnableOption "ceranaNodeCoordinator";

  config = mkIf cfg.enable {
    systemd.services.ceranaNodeCoordinator = {
      description = "Cerana Node Coordinator";
      wantedBy = [ "multi-user.target" ];
      after = [ "cerananet.service" ];
      serviceConfig = {
        Type = "simple";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
        Restart = "always";
        RestartSec = "3";
      };
      preStart = ''
        ${pkgs.coreutils}/bin/mkdir -p ${socketdir}
        ${pkgs.coreutils}/bin/mkdir -p ${cfgdir}
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "service_name": "${name}",' >> ${cfgdir}${cfgfile}
                echo '  "request_timeout": 60,' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
