AC_INIT([sqlite-otel-collector], [0.7.0], [https://github.com/RedShiftVelocity/sqlite-otel/issues])
AM_INIT_AUTOMAKE([-Wall -Werror foreign subdir-objects])
AC_PROG_CC
AC_CONFIG_FILES([
 Makefile
])

# Check for Go compiler
AC_ARG_VAR([GO], [Go compiler command])
AC_PATH_PROG([GO], [go], [no])
if test "x$GO" = "xno"; then
    AC_MSG_ERROR([Go compiler not found. Please install Go 1.21 or later.])
fi

# Check Go version
AC_MSG_CHECKING([for Go version])
GO_VERSION=`$GO version | sed 's/.*go\([[0-9]]\+\.[[0-9]]\+\).*/\1/'`
AC_MSG_RESULT([$GO_VERSION])

# Convert version to comparable format
GO_VERSION_MAJOR=`echo $GO_VERSION | cut -d. -f1`
GO_VERSION_MINOR=`echo $GO_VERSION | cut -d. -f2`

if test $GO_VERSION_MAJOR -lt 1 || (test $GO_VERSION_MAJOR -eq 1 && test $GO_VERSION_MINOR -lt 21); then
    AC_MSG_ERROR([Go 1.21 or later is required. Found $GO_VERSION])
fi

# Check for pkg-config
PKG_PROG_PKG_CONFIG

# Check for SQLite3 (system library - optional, we use go-sqlite3)
AC_ARG_WITH([system-sqlite],
    AS_HELP_STRING([--with-system-sqlite], [Use system SQLite3 library]),
    [with_system_sqlite=$withval],
    [with_system_sqlite=no])

if test "x$with_system_sqlite" = "xyes"; then
    PKG_CHECK_MODULES([SQLITE3], [sqlite3 >= 3.7.0], [
        AC_DEFINE([HAVE_SYSTEM_SQLITE3], [1], [Define if system SQLite3 is available])
    ], [
        AC_MSG_WARN([System SQLite3 not found, will use embedded Go SQLite driver])
    ])
fi

# Check for systemd (for service file installation)
AC_ARG_WITH([systemd],
    AS_HELP_STRING([--with-systemd], [Install systemd service file]),
    [with_systemd=$withval],
    [with_systemd=auto])

if test "x$with_systemd" != "xno"; then
    PKG_CHECK_MODULES([SYSTEMD], [systemd], [
        SYSTEMD_SYSTEM_UNIT_DIR=`$PKG_CONFIG --variable=systemdsystemunitdir systemd`
        AC_SUBST([SYSTEMD_SYSTEM_UNIT_DIR])
        with_systemd=yes
    ], [
        if test "x$with_systemd" = "xyes"; then
            AC_MSG_ERROR([systemd requested but not found])
        fi
        with_systemd=no
    ])
fi

AM_CONDITIONAL([HAVE_SYSTEMD], [test "x$with_systemd" = "xyes"])

# Feature flags
AC_ARG_ENABLE([debug],
    AS_HELP_STRING([--enable-debug], [Enable debug build]),
    [enable_debug=$enableval],
    [enable_debug=no])

if test "x$enable_debug" = "xyes"; then
    GO_BUILD_FLAGS="$GO_BUILD_FLAGS -gcflags='-N -l'"
    AC_DEFINE([DEBUG], [1], [Define for debug builds])
else
    GO_BUILD_FLAGS="$GO_BUILD_FLAGS -ldflags='-s -w'"
fi

AC_SUBST([GO_BUILD_FLAGS])

# Cross-compilation support
AC_ARG_VAR([GOOS], [Target operating system for Go cross-compilation])
AC_ARG_VAR([GOARCH], [Target architecture for Go cross-compilation])

# Installation directories
AC_ARG_WITH([config-dir],
    AS_HELP_STRING([--with-config-dir=DIR], [Configuration directory [[SYSCONFDIR/sqlite-otel]]]),
    [configdir=$withval],
    [configdir='${sysconfdir}/sqlite-otel'])
AC_SUBST([configdir])

AC_ARG_WITH([data-dir],
    AS_HELP_STRING([--with-data-dir=DIR], [Data directory [[LOCALSTATEDIR/lib/sqlite-otel]]]),
    [datadir_custom=$withval],
    [datadir_custom='${localstatedir}/lib/sqlite-otel'])
AC_SUBST([datadir_custom])

AC_ARG_WITH([log-dir],
    AS_HELP_STRING([--with-log-dir=DIR], [Log directory [[LOCALSTATEDIR/log]]]),
    [logdir=$withval],
    [logdir='${localstatedir}/log'])
AC_SUBST([logdir])

# Summary
AC_MSG_NOTICE([
Configuration summary:
  Package:                 $PACKAGE_NAME $PACKAGE_VERSION
  Go compiler:             $GO ($GO_VERSION)
  Debug build:             $enable_debug
  System SQLite3:          $with_system_sqlite
  Systemd integration:     $with_systemd
  
  Installation directories:
    Prefix:                $prefix
    Bindir:                $bindir
    Sysconfdir:            $sysconfdir
    Config dir:            $configdir
    Data dir:              $datadir_custom
    Log dir:               $logdir
])

if test "x$with_systemd" = "xyes"; then
    AC_MSG_NOTICE([    Systemd unit dir:      $SYSTEMD_SYSTEM_UNIT_DIR])
fi

AC_OUTPUT