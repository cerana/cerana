{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaL2Coordinator;
  name = "l2-coord";
  cfgdir = "/data/config/";
  cfgfile = "l2-coordinator.json";
  socketdir = "/task-socket/l2-coordinator/";
  port = "8085";
  daemon = "${pkgs.cerana.bin}/bin/coordinator";
in
{
  options.services.ceranaL2Coordinator.enable = mkEnableOption "ceranaL2Coordinator";

  config = mkIf cfg.enable {
    systemd.services.ceranaL2Coordinator = {
      description = "Cerana Layer 2 Coordinator";
      wantedBy = [ "ceranaLayer2.target" ];
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
                echo '  "external_port": ${port},' >> ${cfgdir}${cfgfile}
                echo '  "request_timeout": 60,' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
