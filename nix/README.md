invoke install-overlay.sh

Then run, e.g.:

```
nix-build -E 'with import <nixpkgs> {}; pkgs.ipxe'
nix-build -E 'with import <nixpkgs> {}; pkgs.cerana'
nix-build -E 'with import <nixpkgs> {}; pkgs.cerana-scripts'
```

Still working on porting the live-media build into this tree
