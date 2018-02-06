{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaDhcpProvider;
  name = "dhcp-provider";
  cfgdir = "/data/config/";
  cfgfile = "dhcp-provider.json";
  socketdir = "/task-socket/l2-coordinator/";
  socket = "coordinator/l2-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/dhcp-provider";
in
{
  options.services.ceranaDhcpProvider.enable = mkEnableOption "ceranaDhcpProvider";

  config = mkIf cfg.enable {
    systemd.services.ceranaDhcpProvider = {
      description = "Cerana DHCP Provider";
      wantedBy = [ "ceranaLayer2.target" ];
      requires = [ "ceranaL2Coordinator.service" "ceranaKvProvider.service" ];
      after = [ "ceranaL2Coordinator.service" "ceranaKvProvider.service" ];
      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = "3";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile}";
      };
      preStart = ''
        rm -f ${socketdir}response/${name}.sock
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                dns=`grep nameserver /etc/resolv.conf | cut -d ' ' -f 2`
                gw=`${pkgs.nettools}/bin/route -n | grep UG | tr -s ' ' | cut -d ' ' -f 2`
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
