{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaNamespaceProvider;
  name = "namespace-provider";
  cfgdir = "/data/config/";
  cfgfile = "namespace-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/namespace-provider";
in
{
  options.services.ceranaNamespaceProvider.enable = mkEnableOption "ceranaNamespaceProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaNamespaceProvider = {
      description = "Cerana Namespace Provider";
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
                echo '  "request_timeout": 5' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
