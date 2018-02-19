{ stdenv, lib, buildGoPackage, fetchFromGitHub }:

buildGoPackage rec {
  name = "godocdown-${version}";
  version = "0bfa0490548148882a54c15fbc52a621a9f50cbe";
  rev = "${version}";

  goPackagePath = "github.com/robertkrimen/godocdown";

  subPackages = [ "godocdown" ];

  src = fetchFromGitHub {
    inherit rev;
    owner = "robertkrimen";
    repo = "godocdown";
    sha256 = "0jgxv2y2anca4xp47lv4lv5n5dcfnfg78bn36vaqvg4ks2gsw0g6";
  };
}
