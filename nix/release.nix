{ pkgs ? import <nixpkgs> {}
}:

let
  version = "1.0";
in rec {
  ipxe = pkgs.ipxe;
  cerana = pkgs.cerana;
  cerana-scripts = pkgs.cerana-scripts;
}
