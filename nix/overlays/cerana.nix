self: super:

let

  inherit (super) callPackage;

in
{

 cerana = callPackage ../pkgs/os-specific/linux/cerana {};
 cerana-scripts = callPackage ../pkgs/os-specific/linux/cerana/scripts.nix {};

}
