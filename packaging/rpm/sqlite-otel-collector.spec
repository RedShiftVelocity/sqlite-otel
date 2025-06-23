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
Source1:        sqlite-otel-collector.sysusers
Source2:        sqlite-otel-collector.yaml

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
export GOPROXY=direct
export GOSUMDB=off
go build -trimpath -ldflags "-s -w -X main.Version=v%{version} -X main.BuildTime=$(date -u '+%%Y-%%m-%%d_%%H:%%M:%%S')" -o sqlite-otel .

%install
# Install binary (rename from sqlite-otel to sqlite-otel-collector)
install -D -m 0755 sqlite-otel %{buildroot}%{_bindir}/%{name}

# Install systemd service file  
install -D -m 0644 packaging/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

# Create directories
install -d -m 0755 %{buildroot}%{_sharedstatedir}/%{name}

# Install sysusers file for user/group creation
install -D -m 0644 %{SOURCE1} %{buildroot}%{_sysusersdir}/%{name}.conf

# Install default configuration file
install -D -m 0640 %{SOURCE2} %{buildroot}%{_sysconfdir}/%{name}/config.yaml

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%{_sysusersdir}/%{name}.conf
%dir %{_sysconfdir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}/config.yaml
%dir %attr(0755, sqlite-otel, sqlite-otel) %{_sharedstatedir}/%{name}

%changelog
* Thu Jun 20 2024 Manish Sinha <manishsinha.tech@gmail.com> - 0.7.0-1
- Initial RPM package
- Cross-platform build support
- Execution logging with rotation
- SQLite-only storage