## ADDED Requirements

### Requirement: File metadata configuration
The system SHALL read file metadata from `data/metadata.yaml`, which defines visibility levels, display names, descriptions, icons, and tags for files and directories.

#### Scenario: Load metadata on startup
- **WHEN** the API server starts
- **THEN** it reads `data/metadata.yaml` and indexes all file metadata entries by path

#### Scenario: Metadata-driven tree display
- **WHEN** the file tree is rendered for a path that has a metadata entry
- **THEN** the display name, icon, and tags from the metadata entry are used instead of the raw file name

### Requirement: Visibility levels
Each file or directory SHALL have a visibility level of `public`, `protected`, or `hidden`, configurable in `metadata.yaml`. The `password.yaml` protection rules SHALL take precedence over metadata visibility.

#### Scenario: Public file visible to all
- **WHEN** a file is marked `public` in metadata
- **THEN** all users can see and read the file regardless of authentication status

#### Scenario: Protected file visible but restricted
- **WHEN** a file is marked `protected` in metadata
- **THEN** unauthenticated users can see the file in the tree but content access requires authentication

#### Scenario: Hidden file excluded
- **WHEN** a file is marked `hidden` in metadata
- **THEN** the file is excluded from tree listings and API responses for unauthenticated users

### Requirement: Category metadata
The system SHALL support category-level metadata in `metadata.yaml` that defines display name, icon, description, and color for each top-level config type directory.

#### Scenario: Category card display
- **WHEN** the home page renders category cards
- **THEN** each card shows the display name, icon, description, and file count from the metadata entry for that category

Example metadata.yaml:
```yaml
categories:
  proxy:
    name: "代理配置"
    icon: "shield"
    description: "Surge / Mihomo 代理规则与节点"
    color: "#4A90D9"
  vim:
    name: "Vim 配置"
    icon: "code"
    description: "Vim / Neovim 编辑器配置"
    color: "#019733"

files:
  "proxy/surge/nodes":
    visibility: hidden
    description: "代理节点文件（需要密码）"
  "proxy/surge/macOS.conf":
    visibility: protected
    description: "macOS Surge 配置"
    tags: ["surge", "macos"]
```