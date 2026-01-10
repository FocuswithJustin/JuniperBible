# usfm2osis

Python tool to convert USFM (Unified Standard Format Markers) files to OSIS XML.

## Installation

### From PyPI

```bash
pip install usfm2osis
```

### NixOS / Nix

```bash
# Using the included derivation
nix-build nixos/

# Or in a shell
nix-shell nixos/
```

### From Source

```bash
cd src/
python -m venv venv
source venv/bin/activate
pip install -e .
```

## Usage

```bash
# Convert a single file
usfm2osis -o output.osis GEN.usfm

# Convert entire directory
usfm2osis -o bible.osis ./usfm_files/

# With specific options
usfm2osis -o output.osis --lang en --id KJV ./usfm/
```

## Official Repository

https://github.com/chrislit/usfm2osis
