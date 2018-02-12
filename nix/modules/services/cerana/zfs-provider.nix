{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaZfsProvider;
  name = "zfs-provider";
  cfgdir = "/data/config/";
  cfgfile = "zfs-provider.json";
  socketdir = "/task-socket/node-coordinator/";
  socket = "coordinator/node-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/zfs-provider";
in
{
  options.services.ceranaZfsProvider.enable = mkEnableOption "ceranaZfsProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaZfsProvider = {
      description = "Cerana ZFS Provider";
      path = [ pkgs.zfs ];
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
                echo '  "coordinator_url": "unix://${socketdir}${socket}"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
