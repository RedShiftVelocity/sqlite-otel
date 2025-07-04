# Development Journal

## [2025-06-23] - PR #86: Ubuntu PPA Package Repository Support
### Actions:
- Created comprehensive Ubuntu PPA packaging structure in debian/ directory
- Implemented debian/control with proper dependencies and package metadata
- Created debian/rules with CGO-enabled Go build process and systemd integration
- Added debian/changelog with version history for Ubuntu 22.04 LTS (Jammy)
- Implemented debian/copyright with proper license attribution
- Created maintainer scripts (postinst, prerm, postrm) for user/service management
- Built automated PPA build and upload script (scripts/build-ppa.sh)
- Added comprehensive PPA documentation and user installation guide

### Decisions:
- Support Ubuntu 22.04 LTS (Jammy Jellyfish)
- Use native source package format (3.0 native) for PPA builds
- Enable CGO for full SQLite support in PPA builds
- Create dedicated sqlite-otel system user with security hardening
- Implement comprehensive systemd service integration
- Provide both automated and manual build processes
- Include detailed troubleshooting and configuration documentation

### Challenges:
- Ensuring CGO-enabled builds work correctly in Launchpad build environment
- Managing version numbering across multiple Ubuntu releases
- Creating proper maintainer scripts for clean installation/removal
- Balancing package complexity with Ubuntu packaging standards
- Coordinating systemd service management with package lifecycle

### Learnings:
- Ubuntu PPA packaging requires source packages, not binary packages
- debian/rules file controls the entire build process including Go compilation
- Launchpad builds packages for multiple architectures automatically
- Version suffixes (~focal1, ~jammy1) enable distribution-specific builds
- GPG signing is mandatory for PPA uploads
- debhelper-compat level 13 provides modern packaging features
- systemd integration requires specific packaging patterns (dh-systemd)

### PPA Structure:
- **Package Name**: sqlite-otel-collector
- **Architectures**: amd64, arm64, armhf
- **Dependencies**: golang-go, libsqlite3-dev, systemd, adduser
- **Installation Path**: /usr/bin/sqlite-otel-collector
- **Service**: sqlite-otel-collector.service
- **Data Directory**: /var/lib/sqlite-otel-collector/
- **System User**: sqlite-otel with security restrictions

## [2025-06-22] - PR #85: Homebrew Package Manager Support
### Actions:
- Created comprehensive Homebrew formula for multi-architecture support
- Added support for macOS (Intel/Apple Silicon) and Linux (x86_64/ARM64/ARM)
- Implemented Homebrew service management with automatic data directory setup
- Added detailed installation instructions and documentation
- Created validation script for formula testing
- Updated main README.md with Homebrew installation section

### Decisions:
- Use direct formula installation from GitHub (no custom tap initially)
- Support all available architectures with platform-specific binary selection
- Include comprehensive service management with brew services
- Provide detailed caveats and usage instructions for users
- Create separate homebrew/ directory for all Homebrew-related files

### Challenges:
- Managing multiple architecture downloads in single formula
- Ensuring proper SHA256 checksums for all platform binaries
- Creating comprehensive service configuration for different platforms
- Balancing simplicity vs comprehensive feature coverage

### Learnings:
- Homebrew formulas can elegantly handle multi-platform binary distribution
- Service blocks in Homebrew provide excellent background service management
- Direct GitHub formula installation is viable for projects without custom taps
- Comprehensive caveats section greatly improves user experience
- Platform detection in Ruby is straightforward with Hardware::CPU methods

## [2025-06-22] - PR #84: Go Module Dependency Management Integration
### Actions:
- Integrated `go mod tidy` verification into CircleCI build pipeline
- Updated Makefile to run `go mod tidy` before all build targets (build, build-all, build-linux)
- Added automated dependency cleanliness checks in CI/CD workflow
- Fixed existing go.mod issue where dependency was incorrectly marked as `// indirect`
- Tested changes locally with successful builds and test execution

### Decisions:
- Follow Go's official publishing guidelines by enforcing tidy modules
- Fail CI builds if go.mod or go.sum files are not properly maintained
- Integrate tidy checks early in the build process (after setup_go_modules)
- Make dependency tidiness a prerequisite for all build operations

### Challenges:
- Existing go.mod had sqlite3 dependency incorrectly marked as indirect
- Need to balance build speed vs. dependency verification overhead
- Ensuring CI/CD fails fast when modules are not tidy

### Learnings:
- `go mod tidy` catches subtle dependency management issues automatically
- Integrating dependency checks into CI prevents publishing untidy modules
- Go publishing best practices require automated enforcement for reliability
- Module tidiness is essential for proper dependency management in published libraries

## [2025-06-20] - v0.7 CircleCI Configuration
### Actions:
- Updated existing CircleCI configuration from basic template to comprehensive CI/CD pipeline
- Added test job with go vet, race detection, and coverage reports
- Added build job for multi-platform builds
- Added release job for version tags
- Implemented Go module caching
- Added proper error handling with set -e

### Decisions:
- Use Go 1.21 Docker image (cimg/go:1.21)
- Run static analysis (go vet) before tests
- Store test results and coverage artifacts
- Build all platforms defined in Makefile
- Trigger release workflow only on version tags (v*)
- Use binaries directory for cleaner artifact management

### Challenges:
- None - leveraged existing CircleCI branch setup

### Learnings:
- CircleCI's caching system speeds up Go module downloads
- Workspace persistence enables artifact sharing between jobs
- set -e ensures fail-fast behavior in shell scripts

### CI/CD Features:
- Automatic testing on all branches with coverage reports
- Static analysis with go vet
- Multi-platform builds via Makefile targets
- Automated release archive creation for tags
- Proper job dependencies and filtering

### Code Review Improvements:
- Created reusable command for Go module setup to reduce duplication
- Pinned Go Docker image to specific version (1.21.6) for reproducibility
- Added explicit binary existence check after build
- Added CIRCLE_TAG validation in release job
- Fixed workflow version to match config version (2.1)

### Additional Improvements (from Gemini review):
- Removed redundant ls commands from build logs
- Extracted release archive creation logic to scripts/create-release-archives.sh
- Fixed test result reporting to preserve go test exit codes (using PIPESTATUS)
- Added go-junit-report for proper JUnit XML test results in CircleCI
- Confirmed test-before-build order is correct for Go projects

## [2025-06-20] - PR #48: v0.6 SQLite-Only Storage
### Actions:
- Removed stdout/file output functionality from handlers
- Deleted handlers/common.go containing WriteTelemetryData function
- Updated handler_common.go to only store data in SQLite
- Updated handler_common.go to use structured logging package
- Replaced all log.Printf calls with appropriate logging levels
- Changed logging to show only execution metadata, not telemetry data content

### Decisions:
- Simplified architecture by removing dual storage mechanism
- All telemetry data now persists only to SQLite database
- Execution logging continues to stdout and file (when configured)
- Telemetry data content is not logged, only metadata (size, type, etc.)

### Challenges:
- Needed to update logging calls after merging with v0.5 logging implementation

### Learnings:
- Removing features can improve code clarity and maintainability
- SQLite provides sufficient storage capabilities for telemetry data
- Clear separation between execution logging and telemetry storage

### Code Review Feedback (Gemini and O3):
- Removal is correct and complete
- Identified potential SQLite write contention under high load (pre-existing issue)
- Suggested future improvements: queuing mechanism for writes, structured logging
- These improvements are beyond v0.6 scope but noted for future work

### Code Review Fixes Applied:
- **HIGH**: Fixed memory usage by switching from io.ReadAll to streaming JSON decoder
- **LOW**: Moved defer r.Body.Close() immediately after MaxBytesReader for safety
- Updated logging to use Content-Length header instead of len(body)
- Created GitHub issues for remaining improvements (async writes, configurable limits, log consolidation)

## [2025-06-20] - PR #47: v0.5 Execution Logging Implementation
### Actions:
- Added logging package with structured logging capabilities
- Added --log-file command-line flag with intelligent defaults
- Implemented execution metadata logging (startup, shutdown, errors)
- Added telemetry activity logging for debugging
- Logs write to both stdout and file when log file is specified

### Decisions:
- Used XDG_STATE_HOME for user mode logs (~/.local/state/sqlite-otel/execution.log)
- Service mode defaults to /var/log/sqlite-otel-collector.log
- Multi-writer approach: logs go to both stdout and file for visibility
- Structured logging with levels: INFO, ERROR, DEBUG
- Thread-safe logging with mutex protection

### Challenges:
- Ensuring thread-safe concurrent logging
- Choosing appropriate default paths for different execution contexts

### Learnings:
- XDG_STATE_HOME is the proper location for application state/logs in user mode
- io.MultiWriter allows efficient writing to multiple destinations
- sync.Once ensures safe one-time initialization in concurrent environments
- os.Getuid() is not portable to Windows - use os.UserHomeDir() for detection
- Pre-initializing global logger prevents race conditions during startup

### Code Review Improvements (from Gemini and O3):
- **CRITICAL**: Replaced os.Getuid() with portable service detection using os.UserHomeDir()
- **HIGH**: Fixed race condition by pre-initializing global logger
- **MEDIUM**: Removed redundant logging before os.Exit()
- **MEDIUM**: Consolidated logging methods to reduce code duplication
- Used internal log() method to eliminate repeated code in Info/Error/Debug

### Second Review Improvements:
- **HIGH**: Fixed potential race on shutdown - logger now resets to stdout after close
- All previous issues confirmed resolved by both reviewers

## [2025-06-20] - PR #46: v0.4 Systemd Service Implementation
### Actions:
- Created systemd service file with security hardening options
- Updated main.go to detect when running as a service and use /var/lib/sqlite-otel-collector/ for database storage
- Created install-service.sh script for one-command installation
- Service runs as dedicated sqlite-otel system user for security

### Decisions:
- Used INVOCATION_ID environment variable to detect systemd execution context
- Service defaults to /var/lib/sqlite-otel-collector/otel-collector.db when running as service
- Can be overridden with explicit --db-path argument if needed
- Service type is "simple" for straightforward process management
- Enabled automatic restart with 5-second delay on failure
- Used journald for logging (StandardOutput=journal)
- Applied security hardening: NoNewPrivileges, PrivateTmp, ProtectSystem=strict

### Challenges:
- Initial approach used brittle service detection logic
- Fixed by making database path explicit in service configuration

### Learnings:
- systemd sets INVOCATION_ID for all service executions
- ProtectSystem=strict requires explicit ReadWritePaths for writable directories
- System users should use --no-create-home and /bin/false shell for security
- Defer statements don't execute on log.Fatalf() - need proper error handling pattern
- IdleTimeout in http.Server helps protect against slowloris attacks
- Build flags -ldflags="-s -w" strip debugging info for smaller binaries
- RestrictAddressFamilies and CapabilityBoundingSet provide additional systemd hardening

### Code Review Improvements (from Gemini and O3):
- **CRITICAL**: Refactored main() to use run() pattern ensuring database cleanup on all exit paths
- Added proper error handling with fmt.Errorf instead of log.Fatalf
- Added IdleTimeout: 120s to HTTP server configuration
- Added type assertion check for TCP listener address
- Enhanced systemd security with RestrictAddressFamilies and empty CapabilityBoundingSet
- Added -ldflags="-s -w" to go build for optimized binary size

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
