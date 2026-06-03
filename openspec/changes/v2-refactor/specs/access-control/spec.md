## ADDED Requirements

### Requirement: Password configuration
The system SHALL read password and file protection rules from `data/password.yaml`. The file SHALL define a single password string and a list of protected paths with protection mode (`hidden` or `masked`).

#### Scenario: Read password config
- **WHEN** the API server starts
- **THEN** it reads `data/password.yaml` and loads the password and protected path rules into memory

#### Scenario: Hot reload password config
- **WHEN** `data/password.yaml` is modified on disk
- **THEN** the server reloads the configuration within 30 seconds without requiring a restart

### Requirement: Single password authentication
The system SHALL support a single password for accessing protected files. The password SHALL be compared using constant-time comparison to prevent timing attacks.

#### Scenario: Password verification
- **WHEN** a user submits a password via `/api/auth`
- **THEN** the system compares it using constant-time comparison against the configured password

### Requirement: Hidden mode protection
When a path is configured with `mode: hidden`, the system SHALL completely exclude the file or directory from API responses for unauthenticated users.

#### Scenario: Hidden directory excluded from tree
- **WHEN** an unauthenticated user requests `GET /api/tree` and a directory is marked `hidden`
- **THEN** that directory and all its children are absent from the response

#### Scenario: Hidden file returns 404
- **WHEN** an unauthenticated user requests a `hidden` file
- **THEN** the server returns `404 Not Found` with no indication the file exists

### Requirement: Masked mode protection
When a path is configured with `mode: masked`, the system SHALL include the file in listings but replace sensitive values in the content for unauthenticated users.

#### Scenario: Masked file in tree listing
- **WHEN** an unauthenticated user requests `GET /api/tree` and a file is marked `masked`
- **THEN** the file appears in the tree with a `protected: true` flag and a lock icon indicator

#### Scenario: Masked content delivery
- **WHEN** an unauthenticated user requests the content of a `masked` file
- **THEN** the system returns the file content with values matching password/key patterns replaced by `***`

Specifically, the following patterns SHALL be masked:
- SS/SSR password parameters (`password=xxx`)
- API key/token parameters (`key=xxx`, `token=xxx`, `secret=xxx`)
- Lines starting with `password`, `key`, or `secret` followed by `=`

#### Scenario: Masked file with auth token
- **WHEN** an authenticated user requests the content of a `masked` file
- **THEN** the full unmasked content is returned

### Requirement: JWT token management
The system SHALL issue JWT tokens with a 7-day expiry. The token SHALL be transmitted via the `Authorization: Bearer <token>` header. Invalid or expired tokens SHALL be rejected with `401 Unauthorized`.

#### Scenario: Valid token acceptance
- **WHEN** a request includes a valid JWT token in the `Authorization` header
- **THEN** the request is treated as authenticated, accessing protected content

#### Scenario: Expired token rejection
- **WHEN** a request includes an expired JWT token
- **THEN** the response is `401 Unauthorized` and the client should prompt for password again