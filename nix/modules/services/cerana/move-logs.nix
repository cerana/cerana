{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.ceranaMoveLogs;
  srcdir = "/var";
  destdir = "/data/logs";
  logfiledir = "log";
  rotatemask = "3";
in
{
  options.services.ceranaMoveLogs.enable = mkEnableOption "ceranaMoveLogs";

  config = mkIf cfg.enable {
    systemd.services.ceranaMoveLogs = {
      description = "Cerana Log File Mover";
      wantedBy = [ "multi-user.target" ];
      requires = [ "ceranapool.service" ];
      after = [ "ceranapool.service" ];
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        ExecStart = "${pkgs.systemd}/bin/systemctl restart systemd-journald.service";
      };
      preStart = ''
        # get systemd-journald to fail
        while pkill systemd-journald; do cat /dev/null; done;
        logdir=${destdir}/${logfiledir}
        bootcountfile=$logdir/boots
        if [ -f $bootcountfile ]; then
                n=`cat $bootcountfile`
                rotatedir=$logdir.$n
                if [ -d $rotatedir ]; then
                        rm -rf $rotatedir
                fi
                mv ${destdir}/${logfiledir} $rotatedir
                n=$(( $n + 1 ))
                n=$(( $n & ${rotatemask} ))
        else
                n=0
        fi
        mkdir -p $logdir
        mv ${srcdir}/${logfiledir} ${destdir}
        ln -s ${destdir}/${logfiledir} ${srcdir}/${logfiledir}
        echo $n >$bootcountfile
        exit 0
        '';
    };
  };
}
