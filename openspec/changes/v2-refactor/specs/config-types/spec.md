## ADDED Requirements

### Requirement: Configuration type registration
The system SHALL support any configuration type by simply creating a directory under `data/configs/`. Each directory becomes a category automatically discoverable by the API.

#### Scenario: Auto-discover config types
- **WHEN** the API server starts
- **THEN** it scans `data/configs/` and registers each top-level directory as a configuration type (e.g., `proxy`, `vim`, `git`)

#### Scenario: New config type via directory creation
- **WHEN** a new directory is created under `data/configs/` (e.g., `data/configs/ssh/`)
- **THEN** the type `ssh` automatically appears in the category list on next API request (within 30s cache refresh)

### Requirement: Syntax highlighter mapping
The system SHALL map file extensions to syntax highlighter languages. The mapping SHALL be configurable via `data/metadata.yaml` and include sensible defaults.

#### Scenario: Default extension mapping
- **WHEN** a file has a known extension (.yaml, .yml, .ini, .conf, .json, .sh, .vim, .toml)
- **THEN** the appropriate syntax highlighter is applied automatically

#### Scenario: Custom extension mapping
- **WHEN** `metadata.yaml` defines a custom mapping for a config type (e.g., `proxy.surge.highlight: "ini"`)
- **THEN** files under that type/path use the specified highlighter

#### Scenario: Default extensions
The following default mappings SHALL be included:
- `.yaml`, `.yml` → YAML
- `.ini`, `.conf` → INI
- `.json` → JSON
- `.sh`, `.bashrc`, `.zshrc` → Shell
- `.vimrc`, `.vim` → Vim script
- `.toml` → TOML
- `.list` → Plain text with comment support
- `.md` → Markdown

### Requirement: Config type templates
The system SHALL support per-type README or description files. If a `README.md` exists in a config type directory, it SHALL be displayed as the category description in the Web UI.

#### Scenario: Category README display
- **WHEN** user navigates to a config type (e.g., `proxy/`)
- **THEN** if `data/configs/proxy/README.md` exists, its rendered content is shown as the category description