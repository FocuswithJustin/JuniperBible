{ pkgs ? import <nixpkgs> {} }:

# UnRTF - RTF to other formats converter
# Used by: format-rtf

pkgs.unrtf

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];

# For offline use, you can build and copy the result:
# nix-build default.nix -o result
# cp -L result/bin/unrtf /path/to/offline/bin/
