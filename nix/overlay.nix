# From GitHub: mozilla/nixpkgs-mozilla/default.nix.

self: super:

with super.lib;

(foldl' (flip extends) (_: super) [

  (import ./overlays/cerana.nix)
  (import ./overlays/go-packages.nix)
  (import ./overlays/ipxe.nix)

]) self
