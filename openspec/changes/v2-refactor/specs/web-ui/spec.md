## ADDED Requirements

### Requirement: File tree browsing
The system SHALL provide a hierarchical file tree view that displays all configuration files organized by type. The tree SHALL support expanding/collapsing directories, clicking to preview file content, and indicating protected files with a lock icon.

#### Scenario: Browse file tree
- **WHEN** user opens the Web UI
- **THEN** the left panel displays a file tree rooted at `data/configs/`, with directories expandable by click

#### Scenario: Preview file content
- **WHEN** user clicks a file in the tree
- **THEN** the right panel displays the file content with syntax highlighting appropriate to the file type

#### Scenario: Protected file indicator
- **WHEN** a file or directory is marked as `masked` in password.yaml
- **THEN** the file appears in the tree with a lock icon overlay
- **AND** unauthenticated users see masked content (sensitive values replaced with `***`)

#### Scenario: Hidden file not visible
- **WHEN** a file or directory is marked as `hidden` in password.yaml
- **THEN** the file does not appear in the tree for unauthenticated users

### Requirement: Syntax highlighting
The system SHALL render file content with syntax highlighting based on file extension and type. Supported formats SHALL include: INI, YAML, JSON, TOML, Shell, Vim script, and plain text.

#### Scenario: YAML file highlighting
- **WHEN** user previews a `.yaml` or `.yml` file
- **THEN** the content is rendered with YAML syntax highlighting (keys, values, comments distinguished by color)

#### Scenario: INI file highlighting
- **WHEN** user previews a `.conf` or `.ini` file
- **THEN** the content is rendered with INI syntax highlighting (sections, keys, values distinguished)

#### Scenario: Unknown file type
- **WHEN** user previews a file with no recognized extension
- **THEN** the content is rendered as plain text with monospace font and line numbers

### Requirement: Copy path
The system SHALL provide a one-click button to copy the full access URL of any file to the clipboard.

#### Scenario: Copy file URL
- **WHEN** user clicks the "Copy URL" button on a file preview
- **THEN** the full URL (e.g., `https://list.magichub.top/api/raw/configs/proxy/surge/Macmini.conf`) is copied to the clipboard

#### Scenario: Copy relative path
- **WHEN** user clicks the "Copy Path" button on a file preview
- **THEN** the relative path from configs root (e.g., `proxy/surge/Macmini.conf`) is copied to the clipboard

### Requirement: Search
The system SHALL provide search functionality with file name search and file content search.

#### Scenario: Search by file name
- **WHEN** user types a keyword in the search bar
- **THEN** all files whose name contains the keyword are listed as search results

#### Scenario: Search by content
- **WHEN** user types a keyword in the search bar and selects "content" mode
- **THEN** all files containing the keyword in their content are listed, with matching lines highlighted

### Requirement: Category view
The system SHALL provide a category-based view that groups files by configuration type (proxy, vim, git, shell, etc.), with each category showing an icon, name, and file count.

#### Scenario: View by category
- **WHEN** user opens the Web UI
- **THEN** the home page displays category cards for each top-level directory under `data/configs/` (e.g., "Proxy", "Vim", "Git")

#### Scenario: Navigate into category
- **WHEN** user clicks a category card
- **THEN** the file tree filters to show only files within that category and the category detail page opens