ACLOCAL_AMFLAGS = -I m4

# Binary name and source files
bin_PROGRAMS = sqlite-otel-collector
sqlite_otel_collector_SOURCES = \
	main.go \
	main_test.go \
	database/db.go \
	database/logs.go \
	database/metrics.go \
	database/shared.go \
	database/traces.go \
	handlers/handler_common.go \
	handlers/logs.go \
	handlers/metrics.go \
	handlers/traces.go \
	logging/logger.go \
	tools/read_telemetry.go

# Go build rules
sqlite-otel-collector$(EXEEXT): $(sqlite_otel_collector_SOURCES) go.mod go.sum
	$(GO) build $(GO_BUILD_FLAGS) -o $@ .

# Custom rules for Go modules
go.mod go.sum:
	@echo "Go module files exist"

# Clean targets
CLEANFILES = sqlite-otel-collector$(EXEEXT)
DISTCLEANFILES = 

# Install additional files
if HAVE_SYSTEMD
systemdunitdir = $(SYSTEMD_SYSTEM_UNIT_DIR)
systemdunit_DATA = sqlite-otel-collector.service
endif

# Configuration and script files
dist_pkgdata_DATA = \
	README.md \
	LICENSE

# Install scripts
dist_bin_SCRIPTS = install-service.sh

# Test targets
check-local:
	$(GO) test ./...

# Build variants for distribution
build-all-local:
	GOOS=linux GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-linux-amd64$(EXEEXT) .
	GOOS=linux GOARCH=arm64 $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-linux-arm64$(EXEEXT) .
	GOOS=linux GOARCH=arm $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-linux-arm$(EXEEXT) .
	GOOS=darwin GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-darwin-amd64$(EXEEXT) .
	GOOS=darwin GOARCH=arm64 $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-darwin-arm64$(EXEEXT) .
	GOOS=windows GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o sqlite-otel-collector-windows-amd64.exe .

# Documentation
dist_man_MANS = 
dist_doc_DATA = README.md INSTALL NEWS AUTHORS

# Create required directories during install
install-data-hook:
	$(MKDIR_P) $(DESTDIR)$(configdir)
	$(MKDIR_P) $(DESTDIR)$(datadir_custom)
	$(MKDIR_P) $(DESTDIR)$(logdir)
	@echo ""
	@echo "sqlite-otel-collector installation complete!"
	@echo ""
	@echo "Configuration directory: $(configdir)"
	@echo "Data directory:          $(datadir_custom)"  
	@echo "Log directory:           $(logdir)"
	@echo ""
if HAVE_SYSTEMD
	@echo "Systemd service installed to: $(SYSTEMD_SYSTEM_UNIT_DIR)"
	@echo "Enable with: systemctl enable sqlite-otel-collector"
	@echo "Start with:  systemctl start sqlite-otel-collector"
	@echo ""
endif
	@echo "Run 'sqlite-otel-collector --help' for usage information."

# Uninstall hook
uninstall-hook:
if HAVE_SYSTEMD
	@echo ""
	@echo "Remember to stop and disable the systemd service:"
	@echo "  systemctl stop sqlite-otel-collector"
	@echo "  systemctl disable sqlite-otel-collector"
	@echo ""
endif

# Distribution extras
EXTRA_DIST = \
	go.mod \
	go.sum \
	sqlite-otel-collector.service \
	test-commands.sh \
	scripts/create-release-archives.sh \
	tools/tools \
	DevJournal.md \
	V1_RELEASE_CHECKLIST.md \
	CLAUDE.md \
	homebrew-publishing-guide.md

# Phony targets
.PHONY: build-all-local check-local

# Override default targets to use Go
all-local: sqlite-otel-collector$(EXEEXT)

# Make sure Go build is executed properly
BUILT_SOURCES = 
SUFFIXES = .go

# Clean Go build cache
clean-local:
	$(GO) clean -cache
	rm -f sqlite-otel-collector-*

# For RPM/DEB packaging
dist-hook:
	find $(distdir) -name '.git*' -exec rm -rf {} \; 2>/dev/null || true