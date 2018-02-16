{ stdenv, lib, buildGoPackage, git, glide10, libseccomp, pkgconfig, fetchFromGitHub }:
let
  srcDef = builtins.fromJSON (builtins.readFile ./cerana.json);
in buildGoPackage rec {
  rev = "${builtins.substring 0 10 srcDef.date}-${builtins.substring 0 10 srcDef.rev}";
  name = "cerana-${rev}";

  goPackagePath = "github.com/cerana/cerana";

  src = fetchFromGitHub {
    owner = "cerana";
    repo = "cerana";
    inherit (srcDef) rev sha256;
  };

  CGO_CFLAGS_ALLOW = "-fms-extensions";

  preConfigure = ''
    export GIT_SSL_CAINFO=/etc/ssl/certs/ca-certificates.crt
    glide install
  '';
  postBuild = "rm $NIX_BUILD_TOP/go/bin/zfs";

  buildInputs = [ git glide10 libseccomp pkgconfig ];
}
