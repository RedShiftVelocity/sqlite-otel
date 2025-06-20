Name:           sqlite-otel-collector
Version:        0.7.0
Release:        1%{?dist}
Summary:        OpenTelemetry collector with SQLite storage
License:        MIT
URL:            https://github.com/RedShiftVelocity/sqlite-otel
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  systemd-rpm-macros
Requires:       systemd

%description
A standalone OpenTelemetry collector service that receives telemetry data
and persists it to an embedded SQLite database. Supports traces, metrics,
and logs via OTLP/HTTP protocol.

%prep
%autosetup

%build
# Build the binary
make build

%install
# Install binary
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}

# Install systemd service file
install -D -m 0644 packaging/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

# Create directories
install -d -m 0755 %{buildroot}%{_localstatedir}/lib/%{name}
install -d -m 0755 %{buildroot}%{_localstatedir}/log

# Install default config (if exists)
if [ -f packaging/config/%{name}.conf ]; then
    install -D -m 0644 packaging/config/%{name}.conf %{buildroot}%{_sysconfdir}/%{name}/%{name}.conf
fi

%pre
# Create user and group with error handling
getent group sqlite-otel >/dev/null || groupadd -r sqlite-otel || {
    echo "Failed to create group sqlite-otel" >&2
    exit 1
}
getent passwd sqlite-otel >/dev/null || \
    useradd -r -g sqlite-otel -d %{_localstatedir}/lib/%{name} -s /sbin/nologin \
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
%systemd_postun_with_restart %{name}.service

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%dir %attr(0755, sqlite-otel, sqlite-otel) %{_localstatedir}/lib/%{name}
%if 0%{?_sysconfdir:1}
%dir %{_sysconfdir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.conf
%endif

%changelog
* Thu Jun 20 2024 Claude Code <noreply@anthropic.com> - 0.7.0-1
- Initial RPM package
- Cross-platform build support
- Execution logging with rotation
- SQLite-only storage