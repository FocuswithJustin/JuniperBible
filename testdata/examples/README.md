# Example Capsules

This directory contains example capsules demonstrating Juniper Bible workflows.

## Sample Capsule

`sample.capsule.tar.xz` - A simple capsule containing a single text file.

### Expected Hashes

| Artifact | SHA-256 |
|----------|---------|
| sample | `97c5bc184d5841cca842579dc4712a43236c754e6f29736bff6136789cb918fe` |

### Verification

```bash
# Verify the capsule
./capsule verify testdata/examples/sample.capsule.tar.xz

# Run selfcheck
./capsule selfcheck testdata/examples/sample.capsule.tar.xz

# Export and verify round-trip
./capsule export testdata/examples/sample.capsule.tar.xz --artifact sample --out /tmp/sample.txt
diff testdata/fixtures/inputs/sample.txt /tmp/sample.txt
```

## SWORD Module Testing

The `testdata/fixtures/sword-test` directory contains a minimal SWORD module structure for testing.

### Structure

```
sword-test/
  mods.d/
    kjv.conf           # Module configuration
  modules/
    texts/ztext/kjv/
      nt.bzz           # Compressed Bible data
```

### Tool Plugin Test

```bash
# Run list-modules profile
./plugins/tool-libsword/tool-libsword run \
  --request request.json \
  --out /tmp/sword-out

# Where request.json contains:
{
  "profile": "list-modules",
  "sword_path": "testdata/fixtures/sword-test"
}
```

## Golden Hashes

Golden hashes are stored in `testdata/goldens/`:

- `sample-artifact.sha256` - Expected hash for sample.txt artifact

## Adding New Examples

1. Create test data in `testdata/fixtures/`
2. Ingest into a capsule: `./capsule ingest <path> --out testdata/examples/<name>.capsule.tar.xz`
3. Verify: `./capsule verify testdata/examples/<name>.capsule.tar.xz`
4. Save golden: `./capsule golden save ... --out testdata/goldens/`
5. Document in this README
