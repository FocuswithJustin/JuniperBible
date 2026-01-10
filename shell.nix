{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go development
    go
    gopls
    gotools
    go-tools
    delve

    # CGO dependencies (for legacy juniper build)
    sqlite
    pkg-config
    gcc

    # Reference tools
    sword

    # Build and test utilities
    gnumake
    git
    jq
    cloc

    # Documentation
    pandoc

    # Archive handling
    zip
    unzip
    gnutar
    xz

    # XML tools (for format plugins)
    libxml2
    libxslt

    # E-book tools (for calibre plugin)
    calibre

    # Network tools (for repoman testing)
    curl
    wget
  ];

  # CGO configuration for sqlite
  CGO_ENABLED = "1";

  shellHook = ''
    export TZ=UTC
    export LC_ALL=C.UTF-8
    export LANG=C.UTF-8

    # Go module configuration
    export GOFLAGS="-mod=readonly"

    echo ""
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║              Juniper Bible Development Environment              ║"
    echo "╠═══════════════════════════════════════════════════════════════╣"
    echo "║                                                               ║"
    echo "║  Build:                                                       ║"
    echo "║    make build          - Build capsule CLI                    ║"
    echo "║    make plugins        - Build all plugins                    ║"
    echo "║    make all            - Build everything and test            ║"
    echo "║                                                               ║"
    echo "║  Test:                                                        ║"
    echo "║    make test           - Run all tests                        ║"
    echo "║    make test-v         - Run tests with verbose output        ║"
    echo "║    make integration    - Run integration tests                ║"
    echo "║                                                               ║"
    echo "║  Juniper Plugins (113 tests):                                 ║"
    echo "║    cd plugins/juniper && go test ./...                        ║"
    echo "║                                                               ║"
    echo "║  Legacy Juniper (comparison testing):                         ║"
    echo "║    make juniper-legacy     - Build legacy CLI                 ║"
    echo "║    make juniper-legacy-all - Build all legacy binaries        ║"
    echo "║                                                               ║"
    echo "║  Documentation:                                               ║"
    echo "║    make docs           - Generate documentation               ║"
    echo "║                                                               ║"
    echo "║  SWORD tools available:                                       ║"
    echo "║    diatheke -b KJV -k \"Gen 1:1\"                               ║"
    echo "║                                                               ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo ""
  '';
}
