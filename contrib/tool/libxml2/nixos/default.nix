{ pkgs ? import <nixpkgs> {} }:

# libxml2 - XML parsing toolkit (includes xmllint)
# libxslt - XSLT transformation toolkit (includes xsltproc)
# Used by: format-osis, format-usx, format-zefania, format-tei, format-xml, format-dbl

pkgs.buildEnv {
  name = "xml-tools";
  paths = [
    pkgs.libxml2
    pkgs.libxslt
  ];
}

# Usage in NixOS configuration:
# environment.systemPackages = [ (import ./default.nix {}) ];

# For offline use, you can build and copy the result:
# nix-build default.nix -o result
# cp -L result/bin/xmllint /path/to/offline/bin/
# cp -L result/bin/xsltproc /path/to/offline/bin/
