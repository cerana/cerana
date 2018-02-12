{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaKvProvider;
  name = "kv-provider";
  cfgdir = "/data/config/";
  cfgfile = "kv-provider.json";
  socketdir = "/task-socket/l2-coordinator/";
  socket = "coordinator/l2-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/kv-provider";
in
{
  options.services.ceranaKvProvider.enable = mkEnableOption "ceranaKvProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaKvProvider = {
      description = "Cerana Kv Provider";
      wantedBy = [ "ceranaLayer2.target" ];
      requires = [ "ceranaConsul.service" ];
      after = [ "ceranaL2Coordinator.service" "ceranaConsul.service" ];
      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = "3";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
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
