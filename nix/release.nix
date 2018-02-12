{ pkgs ? import <nixpkgs> { overlays = [ (import ./overlay.nix) ]; }
, supportedSystems ? [ "x86_64-linux" ]
}:

let
  makeNetboot = config:
    let
      config_evaled = import "${pkgs.path}/nixos/lib/eval-config.nix" config;
      build = config_evaled.config.system.build;
      kernelTarget = config_evaled.pkgs.stdenv.platform.kernelTarget;
    in
      pkgs.symlinkJoin {
        name="netboot";
        paths=[
          build.netbootRamdisk
          build.kernel
          build.netbootIpxeScript
          build.ceranaGrub2Config
        ];
      };

in rec {

  ipxe = pkgs.ipxe;
  cerana = pkgs.cerana;
  cerana-scripts = pkgs.cerana-scripts;

  minimal_media = makeNetboot {
    system = "x86_64-linux";
    modules = import ./modules/module-list.nix ++ [
      "${pkgs.path}/nixos/modules/profiles/minimal.nix"
      ./modules/cerana/cerana.nix
      ./modules/profiles/cerana-hardware.nix
      ./modules/profiles/ceranaos.nix
    ];
  };

}
