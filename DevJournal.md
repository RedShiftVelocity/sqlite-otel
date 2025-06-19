# Development Journal

## [2025-06-19] - PR #25: Fix Race Conditions in GetOrCreate Functions
### Actions: 
- Replaced RETURNING clause with INSERT ON CONFLICT DO NOTHING + SELECT pattern for SQLite 3.24.0+ compatibility
- Added explicit NULL handling for attributes with default empty map
- Made resource_id and scope_id NOT NULL in metrics table to prevent NULL duplicates in unique index
- Removed unused getOrDefault function
- Implemented code quality improvements from O3-mini and Gemini reviews
- Added getStringFromMap helper function to reduce code duplication
- Added documentation for JSON marshaling behavior
- Wrapped schema creation in transaction for atomicity
- Added error handling for database close operation

### Decisions:
- Used INSERT ON CONFLICT DO NOTHING + SELECT pattern instead of RETURNING for broader SQLite compatibility
- Made all foreign key columns NOT NULL in tables with unique indexes to prevent NULL duplicates
- Used fmt.Printf for database close errors since log package might not be available during shutdown
- Kept json.Marshal for attributes serialization with documentation about key sorting behavior

### Challenges:
- RETURNING clause requires SQLite 3.35.0+ which may not be available in all environments
- NULL values in unique indexes can lead to duplicate entries in SQLite
- Balancing between code duplication and over-abstraction when extracting helper functions

### Learnings:
- SQLite's ON CONFLICT clause (requires 3.24.0+) is more widely supported than RETURNING (requires 3.35.0+)
- NULL values in unique indexes behave differently in SQLite - multiple rows with NULL values can exist
- Go's json.Marshal sorts map keys by default, providing canonical output for database comparisons
- Transaction wrapping for DDL statements ensures atomicity even for CREATE IF NOT EXISTS operations

## [2025-01-19] - PR #2: v0.2 File Output

### Actions:
- Created common.go in handlers/ directory with WriteTelemetryData function
- Implemented append-only file writing to telemetry.jsonl in JSON lines format
- Updated all three handlers (traces.go, metrics.go, logs.go) to use common writer
- Removed unused fmt imports from all handler files
- Successfully built and tested compilation

### Decisions:
- Centralized file writing logic in common.go for code reuse
- Used JSON lines format with structure: {"type": "trace/metrics/logs", "body": "<raw data>"}
- Implemented thread-safe file access using sync.Mutex
- Used sync.Once to ensure file is opened only once
- File opened in append mode (O_APPEND) to preserve existing data
- Basic error handling with logging, but doesn't fail requests on write errors

### Challenges:
- Initial build failed due to unused fmt imports after removing print statements
- Fixed by removing unused imports from all handler files

### Learnings:
- JSON lines format (`.jsonl`) is ideal for append-only telemetry logging
- sync.Once ensures one-time initialization in concurrent environments
- Go's strict unused import checking helps maintain clean code
- File operations should be non-blocking to avoid impacting request handling

## [2025-01-19] - PR #1: v0.1 Foundation

### Commit 1: Initial Hello World implementation

#### Actions: 
- Created main.go with simple Hello World implementation
- Initialized Go module with go.mod file
- Created Makefile with build, run, clean, test, and build-all targets
- Attempted to test build locally (Go not installed on development system)

#### Decisions:
- Used standard Go project structure with main.go at root
- Set module name as github.com/sqlite-otel/sqlite-otel
- Targeted Go 1.21 for compatibility
- Created comprehensive Makefile with cross-platform build support

#### Challenges:
- Go runtime not installed on the development system, preventing local testing
- Build verification will need to be done in environment with Go installed

#### Learnings:
- Always verify development environment has required tools before starting
- Makefile provides good abstraction for build commands
- Cross-platform build targets can be pre-configured for future use

### Commit 2: Implement Basic OTLP/HTTP Receiver

#### Actions:
- Replaced Hello World with functional OTLP/HTTP receiver implementation
- Added HTTP server that defaults to port 4318 (OTLP/HTTP standard) with -port flag for customization
- Implemented three OTLP endpoints: /v1/traces, /v1/metrics, /v1/logs
- Added graceful shutdown handling for Ctrl+C (SIGINT) and SIGTERM signals
- Split handlers into separate files in handlers/ directory
- Updated handlers to print raw incoming data to stdout
- Added Content-Type validation (only application/json supported in v0.1)
- Created test-commands.sh with sample curl commands for testing
- Successfully built and tested compilation

#### Decisions:
- Used standard library only (no external dependencies) for v0.1
- Implemented OTLP/HTTP instead of gRPC for simplicity
- Default port 4318 with flag to override (including 0 for ephemeral)
- Handlers print raw data to stdout for observability
- Strict Content-Type validation for JSON-only support
- Organized code with separate handler files for maintainability
- 5-second graceful shutdown timeout
- Proper signal handling for clean port unbinding

#### Challenges:
- Port 4318 often in use, added helpful error messages with alternatives

#### Learnings:
- net.Listen("tcp", ":0") automatically assigns available ephemeral port
- http.Server.Shutdown() provides clean connection draining
- signal.Notify() allows intercepting OS signals for graceful shutdown
- Content-Type validation returns 415 Unsupported Media Type for non-JSON