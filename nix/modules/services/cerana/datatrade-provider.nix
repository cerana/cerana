{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaDataTradeProvider;
  name = "datatrade-provider";
  cfgdir = "/data/config/";
  cfgfile = "datatrade-provider.json";
  socketdir = "/task-socket/l2-coordinator/";
  socket = "coordinator/l2-coord.sock";
  coord_port = "8080";
  daemon = "${pkgs.cerana.bin}/bin/datatrade-provider";
in
{
  options.services.ceranaDataTradeProvider.enable = mkEnableOption "ceranaDataTradeProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaDataTradeProvider = {
      description = "Cerana Data Trade Provider";
      wantedBy = [ "ceranaLayer2.target" ];
      after = [ "ceranaL2Coordinator.service" ];
      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = "3";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
      };
      preStart = ''
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                dns=`grep nameserver /etc/resolv.conf | cut -d ' ' -f 2`
                gw=`${pkgs.nettools}/bin/route -n | grep UG | tr -s ' ' | cut -d ' ' -f 2`
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "service_name": "${name}",' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}",' >> ${cfgdir}${cfgfile}
                echo '  "coordinator_url": "unix://${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "node_coordinator_port": ${coord_port},' >> ${cfgdir}${cfgfile}
                echo '  "dataset_dir": "data/datasets"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
