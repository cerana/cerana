# From GitHub: mozilla/nixpkgs-mozilla/default.nix.

self: super:

with super.lib;

(foldl' (flip extends) (_: super) [

  (import ./overlays/cerana.nix)
  (import ./overlays/glide10.nix)
  (import ./overlays/ipxe.nix)

]) self
