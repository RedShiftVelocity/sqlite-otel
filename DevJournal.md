# Development Journal

## [2025-06-20] - PR #50: v0.7 CircleCI Configuration
### Actions:
- Created comprehensive CircleCI configuration for CI/CD pipeline
- Added test job with race detection and coverage reports
- Added build job for multi-platform Linux binaries
- Added release job for tagged versions
- Implemented Go module caching for faster builds
- Added README documentation for CircleCI setup

### Decisions:
- Use Go 1.21 Docker image (cimg/go:1.21)
- Run tests with race detection enabled
- Store test results and coverage artifacts
- Build only Linux platforms in CI (most common deployment target)
- Trigger release workflow only on version tags (v*)
- Use workspace persistence between build and release jobs

### Challenges:
- None - CircleCI has excellent Go support

### Learnings:
- CircleCI's caching system significantly speeds up Go builds
- Workspace persistence allows artifact sharing between jobs
- Tag-based workflows enable automated releases

### CI/CD Features:
- Automatic testing on all branches
- Coverage report generation and storage
- Multi-platform builds (linux/amd64, linux/arm64, linux/arm)
- Automated release archive creation for tags
- Proper job dependencies and filtering

## [2025-06-20] - Roadmap Reorganization
### Actions:
- Reordered development roadmap to prioritize Service Mode implementation
- Moved v0.6 Service Mode to v0.4
- Moved v0.4 Execution Logging to v0.5
- Moved v0.5 SQLite Focus to v0.6

### Decisions:
- Prioritized systemd service integration earlier in the roadmap
- Service mode capabilities provide more immediate value for deployment scenarios
- Execution logging and SQLite-only focus can be implemented after service foundation is in place

### Challenges:
- None - straightforward reorganization of planned features

### Learnings:
- Service mode integration provides a stable foundation for subsequent features
- Having service capabilities earlier enables better testing of logging and storage features

## [2025-06-19] - PR #27: OTLP/HTTP Protocol Compliance & Improvements
### Actions:
- Fixed HTTP status codes to return 500 on database errors (was returning 200)
- Implemented Content-Type prefix matching for charset support
- Made file write errors fail the request with proper error response
- Created ProcessTelemetryRequest common function to eliminate code duplication
- Reduced each handler from ~65 lines to ~10 lines

### Decisions:
- Used strings.HasPrefix for Content-Type to support "application/json; charset=utf-8"
- Made all errors fail fast with appropriate HTTP status codes
- Created central processing function to ensure consistent behavior across endpoints
- Kept telemetry type as parameter to maintain clear endpoint separation

### Challenges:
- Balancing between code reuse and maintaining clear endpoint boundaries
- Ensuring all error paths return appropriate HTTP status codes
- Maintaining backward compatibility while improving error handling

### Learnings:
- OTLP/HTTP spec requires proper HTTP status codes for different error conditions
- Content-Type headers often include charset parameters that need prefix matching
- Significant code reduction possible through well-designed common functions
- Go's function parameters can accept other functions, enabling clean abstraction

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

### Deep Dive: Key Technical Decisions

#### 1. Why NOT NULL on Foreign Keys in Unique Indexes
- **Problem**: SQLite treats NULL != NULL in unique constraints, allowing duplicate entries
- **Example**: Could have multiple metrics with same name but NULL resource_id
- **Solution**: Made resource_id and scope_id NOT NULL to enforce true uniqueness
- **OTLP Compliance**: Every metric MUST have a resource and scope per spec

#### 2. Transaction Wrapping for Schema Creation
- **Problem**: Partial schema creation on failure leaves database inconsistent
- **Benefits**:
  - Atomic rollback if any table/index fails
  - Better error recovery (no manual cleanup needed)
  - Performance improvement (single transaction vs multiple)
  - SQLite wraps each statement in implicit transaction anyway
- **Pattern**: defer tx.Rollback() ensures cleanup even on panic

#### 3. Safe Type Extraction with getStringFromMap
- **Problem**: Direct type assertions on map[string]interface{} can panic
- **Real Risk**: External OTLP data might have wrong types (e.g., number instead of string)
- **Solution**: Two-level safety check:
  1. Check key exists and not nil
  2. Use safe type assertion with ok flag
- **Result**: Server stays running even with malformed input

### SQLite Version Requirements
- **Minimum**: SQLite 3.24.0 (released 2018-06-04)
- **Required for**: ON CONFLICT clause support
- **Not using**: RETURNING clause (requires 3.35.0+, released 2021-03-12)
- **Rationale**: Broader compatibility with enterprise Linux distributions

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