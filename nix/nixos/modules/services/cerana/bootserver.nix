{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaBootserver;
  name = "bootserver";
  cfgdir = "/data/config/";
  cfgfile = "bootserver.json";
  socketdir = "/task-socket/l2-coordinator/";
  socket = "coordinator/l2-coord.sock";
  daemon = "${pkgs.cerana.bin}/bin/bootserver";
in
{
  options.services.ceranaBootserver.enable = mkEnableOption "ceranaBootserver";

  config = mkIf cfg.enable {
    systemd.services.ceranaBootserver = {
      description = "Cerana PXE Boot Server";
      wantedBy = [ "ceranaLayer2.target" ];
      after = [ "ceranaL2Coordinator.service" "ceranaDhcpProvider.service" "systemd-networkd.service" "ceranaPlatformImport.service" ];
      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = "3";
        ExecStart = "${daemon} -c ${cfgdir}${cfgfile} --log_level=info";
      };
      preStart = ''
        rm -f ${socketdir}response/${name}.sock
        if [ ! -f ${cfgdir}${cfgfile} ]; then
                echo "{" > ${cfgdir}${cfgfile}
                echo '  "service_name": "${name}",' >> ${cfgdir}${cfgfile}
                echo '  "socket_dir": "${socketdir}",' >> ${cfgdir}${cfgfile}
                echo '  "coordinator_url": "unix://${socketdir}${socket}",' >> ${cfgdir}${cfgfile}
                echo '  "dhcp-interface": "mgmt0",' >> ${cfgdir}${cfgfile}
                echo '  "ipxe": "${pkgs.ipxe}/undionly.kpxe",' >> ${cfgdir}${cfgfile}
                # Can use symlinks for these which may help with updates.
                # Actually, "current" is a symlink.
                echo '  "initrd": "/data/platform/current/initrd",' >> ${cfgdir}${cfgfile}
                echo '  "kernel": "/data/platform/current/bzImage"' >> ${cfgdir}${cfgfile}
                echo "}" >> ${cfgdir}${cfgfile}
        fi
        '';
    };
  };
}
