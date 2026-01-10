# SWORD Project Nix Package
# For offline use, download the source tarball and place in this directory
#
# Usage:
#   nix-build default.nix
#   nix-shell default.nix
#
# To fetch source for offline use:
#   nix-prefetch-url https://www.crosswire.org/ftpmirror/pub/sword/source/v1.9/sword-1.9.0.tar.gz

{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation rec {
  pname = "sword";
  version = "1.9.0";

  src = pkgs.fetchurl {
    url = "https://www.crosswire.org/ftpmirror/pub/sword/source/v1.9/sword-${version}.tar.gz";
    sha256 = "sha256-0gCq0k4RV3jNPM3rKqZR8vwQVxnQHqnI4c9Vr/DxUeM=";
  };

  nativeBuildInputs = with pkgs; [
    cmake
    pkg-config
  ];

  buildInputs = with pkgs; [
    zlib
    icu
    curl
    clucene_core_2
  ];

  cmakeFlags = [
    "-DSWORD_BUILD_TESTS=OFF"
    "-DLIBSWORD_LIBRARY_TYPE=SHARED"
  ];

  meta = with pkgs.lib; {
    description = "Free Bible Software Project - library and utilities";
    homepage = "https://www.crosswire.org/sword/";
    license = licenses.gpl2Plus;
    platforms = platforms.unix;
    maintainers = [];
  };
}
