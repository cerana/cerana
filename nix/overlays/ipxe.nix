self: super:

let

  inherit (super) callPackage;

in
{

 ipxe = callPackage ../pkgs/tools/misc/ipxe {};

}
