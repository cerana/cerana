{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.targets.cerana;
in
{
  options.targets.cerana.enable = mkEnableOption "cerana";

  config = mkIf cfg.enable {
    systemd.targets.cerana = {
      description = "Full Cerana System";
      after = [ "ceranapool.service" "cerananet.service"];
    };
  };
}
