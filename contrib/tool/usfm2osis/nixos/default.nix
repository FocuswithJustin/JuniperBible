# usfm2osis Nix Package
#
# Usage:
#   nix-build default.nix
#   nix-shell default.nix

{ pkgs ? import <nixpkgs> {} }:

pkgs.python3Packages.buildPythonApplication rec {
  pname = "usfm2osis";
  version = "0.7.0";

  src = pkgs.fetchFromGitHub {
    owner = "chrislit";
    repo = "usfm2osis";
    rev = "v${version}";
    sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Update with actual hash
  };

  propagatedBuildInputs = with pkgs.python3Packages; [
    lxml
  ];

  meta = with pkgs.lib; {
    description = "USFM to OSIS XML converter";
    homepage = "https://github.com/chrislit/usfm2osis";
    license = licenses.gpl3;
    platforms = platforms.unix;
  };
}
