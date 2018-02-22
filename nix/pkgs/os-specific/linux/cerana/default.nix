{ stdenv, lib, buildGoPackage, git, glide10, godocdown, libseccomp, pkgconfig, fetchgit }:
let
  srcDef = builtins.fromJSON (builtins.readFile ./cerana.json);
in buildGoPackage rec {
  rev = "${builtins.substring 0 10 srcDef.date}-${builtins.substring 0 10 srcDef.rev}";
  name = "cerana-${rev}";

  goPackagePath = "github.com/cerana/cerana";

  # We intentionally use fetchgit rather than the wrapper for github
  # so that users can build and test a locally comitted tree
  # see update-local.sh
  src = fetchgit {
    postFetch = "echo hello world;";
    inherit (srcDef) url rev sha256;
  };

  CGO_CFLAGS_ALLOW = "-fms-extensions";

  preConfigure = ''
    export GIT_SSL_CAINFO=/etc/ssl/certs/ca-certificates.crt
    glide install
  '';
  postBuild = "rm $NIX_BUILD_TOP/go/bin/zfs";

  buildInputs = [ git glide10 godocdown libseccomp pkgconfig ];
}
