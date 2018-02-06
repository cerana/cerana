{ stdenv, lib, buildGoPackage, git, glide10, libseccomp, pkgconfig, fetchFromGitHub }:

buildGoPackage rec {
  name = "cerana-${version}";
     version = "2016-08-31";
     owner = "cerana";
     repo = "cerana";
     rev = "5c345c6290b7aa1fec8410fe85f7c2466f2df60e";

  goPackagePath = "github.com/cerana/cerana";

  src = fetchFromGitHub {
    owner = "cerana";
    repo = "cerana";
    inherit rev;
    sha256 = "1hf59qcv9p27jbb5zns3bafbwjxyijxlqqakpydhlklk40a4slm9";
  };

  preConfigure = ''
    export GIT_SSL_CAINFO=/etc/ssl/certs/ca-certificates.crt
    glide install
  '';
  postBuild = "rm $NIX_BUILD_TOP/go/bin/zfs";

  buildInputs = [ git glide10 libseccomp pkgconfig ];
}
