#
# NOTE FOR DEVELOPERS:
# This spec file assumes a standard Fedora/RHEL build environment.
# For compatibility on other systems (e.g., Ubuntu CI), any
# non-standard macros used here MUST also be defined in the
# `build-rpm.sh` script.
# Required Macros: %{_unitdir}, %systemd_post, %systemd_preun, %systemd_postun
#

Name:           sqlite-otel-collector
Version:        0.7.0
Release:        1%{?dist}
Summary:        OpenTelemetry collector with SQLite storage
License:        MIT
URL:            https://github.com/RedShiftVelocity/sqlite-otel
Source0:        %{name}-%{version}.tar.gz

# BuildRequires:  golang >= 1.21  # Not needed for pre-built binary packaging
BuildRequires:  systemd-rpm-macros
Requires:       systemd

%description
A standalone OpenTelemetry collector service that receives telemetry data
and persists it to an embedded SQLite database. Supports traces, metrics,
and logs via OTLP/HTTP protocol.

%prep
%autosetup

%build
# Binary should already be built - just verify it exists
test -f sqlite-otel || (echo "Binary sqlite-otel not found. Run 'make build' first." && exit 1)

%install
# Install binary (rename from sqlite-otel to sqlite-otel-collector)
install -D -m 0755 sqlite-otel %{buildroot}/usr/bin/%{name}

# Install systemd service file  
install -D -m 0644 packaging/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

# Create directories
install -d -m 0755 %{buildroot}/var/lib/%{name}
install -d -m 0755 %{buildroot}/var/log

# No default config file for now

%pre
# Create user and group with error handling
getent group sqlite-otel >/dev/null || groupadd -r sqlite-otel || {
    echo "Failed to create group sqlite-otel" >&2
    exit 1
}
getent passwd sqlite-otel >/dev/null || \
    useradd -r -g sqlite-otel -d /var/lib/%{name} -s /sbin/nologin \
    -c "SQLite OTEL Collector" sqlite-otel || {
    echo "Failed to create user sqlite-otel" >&2
    exit 1
}
exit 0

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service

%files
/usr/bin/%{name}
%{_unitdir}/%{name}.service
%dir %attr(0755, sqlite-otel, sqlite-otel) /var/lib/%{name}

%changelog
* Thu Jun 20 2024 Manish Sinha <manishsinha.tech@gmail.com> - 0.7.0-1
- Initial RPM package
- Cross-platform build support
- Execution logging with rotation
- SQLite-only storage