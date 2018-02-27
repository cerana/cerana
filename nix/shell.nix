with import ./nixpkgs { overlays = [ (import ./overlay.nix) ]; };

stdenv.mkDerivation rec {
  name = "cerana-build-env";
  env = buildEnv { name = name; paths = buildInputs; };
  buildInputs = [
    cerana.buildInputs
    which
    less
    godocdown
    ] ++ [ pkgs.nix ];
}
