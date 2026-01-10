# Plugins Directory

This directory is reserved for external and premium plugins.

## Built-in Plugins

Core plugin functionality is embedded directly into the main binaries (`capsule`, `capsule-web`, `capsule-api`). No external plugins are required for standard operation.

## Adding External Plugins

To add external plugins:

1. Create a subdirectory under `format/`, `tool/`, or the appropriate kind directory
2. Include a `plugin.json` manifest file
3. Include the compiled plugin binary

Example structure:
```
plugins/
├── format/
│   └── my-plugin/
│       ├── plugin.json
│       └── format-my-plugin
└── tool/
    └── my-tool/
        ├── plugin.json
        └── tool-my-tool
```

## Enabling External Plugins

External plugins are disabled by default. To enable them, use the `--plugins-external` flag or set the `CAPSULE_PLUGINS_EXTERNAL=1` environment variable.

## Plugin Development

See the main documentation for plugin development guidelines.
