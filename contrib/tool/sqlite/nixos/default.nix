{ pkgs ? import <nixpkgs> {} }:

# SQLite - Lightweight SQL database engine
# Used by: format-sqlite, format-esword

pkgs.sqlite

# For full-featured build with all extensions:
# pkgs.sqlite.override {
#   interactive = true;
# }

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];

# For offline use, you can build and copy the result:
# nix-build default.nix -o result
# cp -L result/bin/sqlite3 /path/to/offline/bin/
