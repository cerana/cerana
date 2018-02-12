{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaServiceProvider;
  name = "service-provider";
  cfgdir = "/data/config/";
  cfgfile = "service-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  rollbackCloneCmd = "/run/current-system/sw/bin/rollback_clone";
  datasetCloneDir = "data/running-clones";
  daemon = "${pkgs.cerana.bin}/bin/service-provider";
in
{
  options.services.ceranaServiceProvider.enable = mkEnableOption "ceranaServiceProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaServiceProvider = {
      description = "Cerana Service Provider";
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
        find ${socketdir} -iname \*${name}.sock -delete
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "service_name": "${name}",' >> ${cfgdir}${cfgfile}
                echo '  "log_level": "debug",' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}",' >> ${cfgdir}${cfgfile}
                echo '  "coordinator_url": "unix://${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "request_timeout": 5,' >> ${cfgdir}${cfgfile}
                echo '  "rollback_clone_cmd": "${rollbackCloneCmd}",' >> ${cfgdir}${cfgfile}
                echo '  "dataset_clone_dir": "${datasetCloneDir}"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
