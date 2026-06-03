## ADDED Requirements

### Requirement: File browsing API
The system SHALL provide a REST API endpoint `GET /api/tree` that returns the hierarchical file structure under `data/configs/`, respecting visibility rules from `password.yaml` (hiding `hidden` items for unauthenticated requests).

#### Scenario: Unauthenticated tree request
- **WHEN** a request is made to `GET /api/tree` without a valid JWT token
- **THEN** the response contains the full file tree with `hidden` items excluded and `masked` items marked with a `protected: true` flag

#### Scenario: Authenticated tree request
- **WHEN** a request is made to `GET /api/tree` with a valid JWT token
- **THEN** the response contains the complete file tree including all items, with `masked` items still marked `protected: true` but content accessible

### Requirement: File content API
The system SHALL provide `GET /api/file/:path` that returns file content. For protected files, unauthenticated requests SHALL return masked content; authenticated requests SHALL return full content.

#### Scenario: Read public file
- **WHEN** an unauthenticated request is made to `GET /api/file/proxy/surge/rules/direct.list`
- **THEN** the full file content is returned with `Content-Type` set appropriately

#### Scenario: Read masked file without auth
- **WHEN** an unauthenticated request is made to `GET /api/file/proxy/surge/nodes/dawang.ini`
- **THEN** the file content is returned with all lines matching password patterns replaced by `***`

#### Scenario: Read masked file with auth
- **WHEN** an authenticated request is made to `GET /api/file/proxy/surge/nodes/dawang.ini`
- **THEN** the full unmasked file content is returned

#### Scenario: Read hidden file without auth
- **WHEN** an unauthenticated request is made to `GET /api/file/proxy/surge/nodes/dawang.ini` and this file is marked `hidden`
- **THEN** the response is `404 Not Found`

### Requirement: Raw file serving with backward compatibility
The system SHALL provide `GET /api/raw/:path` that serves raw file content (without JSON wrapping) with correct MIME type, matching the behavior of the old static file server. The system SHALL also route legacy URLs to their new paths.

#### Scenario: Serve Surge config via legacy URL
- **WHEN** a request is made to `GET /d/surge/Macmini.conf`
- **THEN** the content of `data/configs/proxy/surge/Macmini.conf` is served with `Content-Type: text/plain`

#### Scenario: Serve Mihomo config via legacy URL
- **WHEN** a request is made to `GET /d/clash/config.yaml`
- **THEN** the content of `data/configs/proxy/mihomo/config.yaml` is served with `Content-Type: text/yaml`

#### Scenario: Raw file with password protection
- **WHEN** a request is made to a raw URL for a protected file without authentication
- **THEN** if the file is `masked`, the response contains masked content; if `hidden`, the response is `404 Not Found`

### Requirement: Authentication API
The system SHALL provide `POST /api/auth` that accepts a password and returns a JWT token on success.

#### Scenario: Correct password
- **WHEN** a request is made to `POST /api/auth` with `{"password": "passward"}`
- **THEN** a JWT token is returned with `200 OK` and an expiry of 7 days

#### Scenario: Incorrect password
- **WHEN** a request is made to `POST /api/auth` with `{"password": "wrong"}`
- **THEN** the response is `401 Unauthorized` with no token

### Requirement: Search API
The system SHALL provide `GET /api/search?q=keyword&type=name|content` for file name and content search.

#### Scenario: Search by file name
- **WHEN** a request is made to `GET /api/search?q=Macmini&type=name`
- **THEN** the response lists all files whose name contains "Macmini", with their paths

#### Scenario: Search by file content
- **WHEN** a request is made to `GET /api/search?q=encrypt-method&type=content`
- **THEN** the response lists all files containing "encrypt-method", with matching line numbers and excerpts (masked appropriately for protected files if unauthenticated)