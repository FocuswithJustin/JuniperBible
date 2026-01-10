# osis2mod

Part of the SWORD Project utilities. Converts OSIS XML files into SWORD modules.

## Note

This tool is included with the sword-utils package. See `../sword-utils/` for installation.

## Usage

```bash
# Create a Bible module
osis2mod ./modules/texts/ztext/MyBible/ bible.osis.xml

# With compression
osis2mod ./modules/texts/ztext/MyBible/ bible.osis.xml -z z

# Specify versification
osis2mod ./modules/texts/ztext/MyBible/ bible.osis.xml -v KJV
```

## Options

- `-z <level>` - Compression level (none, z, l, x)
- `-v <system>` - Versification system (KJV, NRSV, MT, etc.)
- `-c <cipher>` - Encryption cipher key
- `-N` - Disable reference normalization
- `-s <scheme>` - Strong's markup scheme

## Creating a Complete Module

1. Convert source to OSIS XML
2. Run osis2mod to create module files
3. Create a .conf file in mods.d/
4. Test with diatheke

Example .conf file:
```ini
[MyBible]
DataPath=./modules/texts/ztext/mybible/
ModDrv=zText
Lang=en
Encoding=UTF-8
SourceType=OSIS
```

## Official Documentation

https://wiki.crosswire.org/DevTools:Modules
