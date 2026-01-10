{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go toolchain for running tests
    go

    # SWORD tools (diatheke, mod2imp, osis2mod)
    sword

    # Calibre e-book tools (ebook-convert, ebook-meta)
    calibre

    # Pandoc document converter
    pandoc

    # libxml2 tools (xmllint, xsltproc)
    libxml2

    # SQLite CLI
    sqlite

    # RTF converter
    unrtf

    # Test utilities
    gnumake
    git
  ];

  shellHook = ''
    export TZ=UTC
    export LC_ALL=C.UTF-8
    export LANG=C.UTF-8

    # Ensure CGO is disabled for pure Go tests
    export CGO_ENABLED=0

    echo ""
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║          Juniper Bible Integration Test Environment            ║"
    echo "╠═══════════════════════════════════════════════════════════════╣"
    echo "║                                                               ║"
    echo "║  Available Tools:                                             ║"
    echo "║    SWORD:    diatheke, mod2imp, osis2mod                      ║"
    echo "║    Calibre:  ebook-convert, ebook-meta                        ║"
    echo "║    Pandoc:   pandoc                                           ║"
    echo "║    XML:      xmllint, xsltproc                                ║"
    echo "║    SQLite:   sqlite3                                          ║"
    echo "║    RTF:      unrtf                                            ║"
    echo "║                                                               ║"
    echo "║  Run Tests:                                                   ║"
    echo "║    go test ./tests/integration/... -v                         ║"
    echo "║    go test ./tests/integration -run TestSWORD -v              ║"
    echo "║    go test ./tests/integration -run TestCalibre -v            ║"
    echo "║    go test ./tests/integration -run TestPandoc -v             ║"
    echo "║    go test ./tests/integration -run TestXML -v                ║"
    echo "║    go test ./tests/integration -run TestSQLite -v             ║"
    echo "║                                                               ║"
    echo "║  Check Tool Availability:                                     ║"
    echo "║    go test ./tests/integration -run TestToolsAvailable -v     ║"
    echo "║                                                               ║"
    echo "║  Or use Makefile:                                             ║"
    echo "║    make test-integration                                      ║"
    echo "║                                                               ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo ""
  '';
}
