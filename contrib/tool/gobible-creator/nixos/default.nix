{ pkgs ? import <nixpkgs> {} }:

# GoBible Creator - J2ME Bible application creator
# Used by: format-gobible
#
# Note: GoBible Creator is a Java application, not packaged in nixpkgs.
# This derivation provides the Java runtime needed to run it.

pkgs.buildEnv {
  name = "gobible-runtime";
  paths = [
    pkgs.jdk  # Java Development Kit to run GoBibleCreator.jar
  ];
}

# To use GoBible Creator:
# 1. Download GoBibleCreator.jar from SourceForge
# 2. Run: java -jar GoBibleCreator.jar collections.txt

# For a full packaging, you could create a wrapper:
# pkgs.writeShellScriptBin "gobible-creator" ''
#   ${pkgs.jdk}/bin/java -jar /path/to/GoBibleCreator.jar "$@"
# ''

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];
