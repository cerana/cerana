self: super:

let

  inherit (super) callPackage;

in
{

 glide10 = callPackage ../pkgs/development/tools/glide/glide10.nix {};

}
