## ADDED Requirements

### Requirement: Directory structure migration
The system SHALL migrate existing configuration files from the current directory structure to the new `data/configs/` layout while preserving all file contents exactly.

#### Scenario: Surge files migration
- **WHEN** migration is executed
- **THEN** all files from `surge/` are copied to `data/configs/proxy/surge/` maintaining relative paths
- **AND** `surge/nodes/*.ini` → `data/configs/proxy/surge/nodes/*.ini`
- **AND** `surge/macOS.conf` → `data/configs/proxy/surge/macOS.conf`
- **AND** `surge/iOS.conf` → `data/configs/proxy/surge/iOS.conf`
- **AND** `surge/Macmini.conf` → `data/configs/proxy/surge/Macmini.conf`
- **AND** `surge/rules/**` → `data/configs/proxy/surge/rules/**`
- **AND** `surge/modules/**` → `data/configs/proxy/surge/modules/**`
- **AND** `surge/scripts/**` → `data/configs/proxy/surge/scripts/**`
- **AND** `surge/assets/**` → `data/configs/proxy/surge/assets/**`

#### Scenario: Mihomo files migration
- **WHEN** migration is executed
- **THEN** all Mihomo config files are copied to `data/configs/proxy/mihomo/`
- **AND** `clash/config.yaml` → `data/configs/proxy/mihomo/config.yaml`
- **AND** `clash/mihomo/config.yaml` → `data/configs/proxy/mihomo/config-nas.yaml`
- **AND** `clash/mihomo/config-android.yaml` → `data/configs/proxy/mihomo/config-android.yaml`

#### Scenario: Vim config migration
- **WHEN** migration is executed
- **THEN** `config/vimrc` is copied to `data/configs/vim/vimrc`

### Requirement: Legacy URL backward compatibility
The system SHALL maintain backward compatibility with existing Surge managed URLs and Mihomo cron URLs through API route mapping.

#### Scenario: Surge managed URL compatibility
- **WHEN** a Surge client requests `GET /d/surge/Macmini.conf`
- **THEN** the server responds with the content of `data/configs/proxy/surge/Macmini.conf` with correct MIME type
- **AND** the `#!MANAGED-CONFIG` URL inside the file content continues to resolve correctly

#### Scenario: Surge nodes URL compatibility
- **WHEN** a Surge client requests `GET /d/surge/nodes/dawang.ini`
- **THEN** the server responds with the full unmasked content (Surge clients need actual node data)
- **AND** this route bypasses password protection for machine-to-machine access

#### Scenario: Mihomo config URL compatibility
- **WHEN** a Mihomo gateway requests `GET /d/clash/config.yaml` or `GET /d/clash/mihomo/config.yaml`
- **THEN** the server responds with the corresponding YAML content

#### Scenario: Clash list URL compatibility
- **WHEN** a client requests `GET /d/clash/list.yaml`
- **THEN** the server responds with the content of `data/configs/proxy/clash-list.yaml`

### Requirement: Migration script
The system SHALL provide a migration script (`scripts/migrate-v1.sh`) that copies files from old paths to new paths and generates initial `password.yaml` and `metadata.yaml`.

#### Scenario: Run migration script
- **WHEN** `bash scripts/migrate-v1.sh` is executed
- **THEN** all core config files are copied to `data/configs/` with correct paths
- **AND** `data/password.yaml` is generated with default protection rules for known sensitive paths
- **AND** `data/metadata.yaml` is generated with category definitions and file visibility rules