{ pkgs ? import <nixpkgs> {} }:

# Calibre - E-book management and conversion toolkit
# Used by: format-epub

pkgs.calibre

# For headless/server use without GUI:
# pkgs.calibre.override {
#   unrarSupport = false;
# }

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];

# For offline use, you can build and copy the result:
# nix-build default.nix -o result
# cp -L result/bin/ebook-convert /path/to/offline/bin/
