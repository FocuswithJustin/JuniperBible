{ pkgs ? import <nixpkgs> {} }:

# Pandoc - Universal document converter
# Used by: format-markdown, format-html, format-epub, format-rtf, format-odf, format-txt

pkgs.pandoc

# For a custom build with specific extensions:
# pkgs.haskellPackages.pandoc.overrideAttrs (old: {
#   # Custom configuration here
# })

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];

# For offline use, you can build and copy the result:
# nix-build default.nix -o result
# cp -L result/bin/pandoc /path/to/offline/bin/
