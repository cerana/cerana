self: super:

let
  inherit (super) callPackage;
in {
  cerana = callPackage pkgs/os-specific/linux/cerana {};
  cerana-scripts = callPackage pkgs/os-specific/linux/cerana/scripts.nix {};
  glide10 = callPackage pkgs/development/tools/glide/glide10.nix {};
  godocdown = callPackage pkgs/development/tools/godocdown {};
}
