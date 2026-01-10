# SWORD Utilities

The SWORD Project utilities for working with SWORD Bible modules.

## Tools Included

- **diatheke** - Command-line Bible access tool
- **mod2osis** - Convert SWORD modules to OSIS XML
- **mod2imp** - Export SWORD modules to IMP format
- **osis2mod** - Create SWORD modules from OSIS XML
- **imp2ld** - Import lexicon/dictionary data

## Installation

### NixOS / Nix

```bash
# Using flakes
nix shell nixpkgs#sword

# Or add to configuration.nix:
environment.systemPackages = [ pkgs.sword ];
```

### Debian/Ubuntu

```bash
apt install libsword-utils
```

### macOS (Homebrew)

```bash
brew install sword
```

### From Source

See `src/` directory for source code.

```bash
cd src/sword-*
./autogen.sh
./configure
make
make install
```

## Usage

```bash
# List installed modules
diatheke -b system -k modulelist

# Render a verse
diatheke -b KJV -k Gen.1.1

# Export module to OSIS
mod2osis KJV > kjv.osis.xml

# Export all keys
mod2imp KJV > kjv.imp

# Create module from OSIS
osis2mod ./output kjv.osis.xml
```

## Official Website

https://www.crosswire.org/sword/
