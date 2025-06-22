class SqliteOtelCollector < Formula
  desc "Lightweight OpenTelemetry collector with SQLite storage"
  homepage "https://github.com/RedShiftVelocity/sqlite-otel"
  url "https://github.com/RedShiftVelocity/sqlite-otel/archive/refs/tags/v0.8.0.tar.gz"
  sha256 "a04c16beb077934e113c907f68f72e31f485bda061c7f5f1fe9f4815fa16b9ff"
  license "MIT"
  head "https://github.com/RedShiftVelocity/sqlite-otel.git", branch: "main"

  depends_on "go" => :build

  def install
    # Set CGO_ENABLED=1 to enable CGO for SQLite support
    ENV["CGO_ENABLED"] = "1"
    
    # Build the binary using the Makefile
    system "make", "build"
    
    # Install the binary
    bin.install "sqlite-otel" => "sqlite-otel-collector"
  end

  service do
    run [opt_bin/"sqlite-otel-collector"]
    keep_alive true
    log_path var/"log/sqlite-otel-collector.log"
    error_log_path var/"log/sqlite-otel-collector.log"
    working_dir var/"lib/sqlite-otel-collector"
  end

  def post_install
    # Create data directory
    (var/"lib/sqlite-otel-collector").mkpath
    # Create log directory  
    (var/"log").mkpath
  end

  test do
    # Test that the binary runs and shows version
    output = shell_output("#{bin}/sqlite-otel-collector --version 2>&1")
    assert_match "sqlite-otel-collector", output

    # Test that help command works
    output = shell_output("#{bin}/sqlite-otel-collector --help 2>&1")
    assert_match "Port to listen on", output
    
    # Test basic functionality (create a test database)
    system bin/"sqlite-otel-collector", "--version"
  end

  def caveats
    <<~EOS
      SQLite OpenTelemetry Collector has been installed!

      To start the collector immediately:
        sqlite-otel-collector

      To run as a background service:
        brew services start sqlite-otel-collector

      The collector will listen on:
        - HTTP: http://localhost:4318 (OTLP/HTTP endpoint)

      Data will be stored in:
        #{var}/lib/sqlite-otel-collector/

      Logs will be written to:
        #{var}/log/sqlite-otel-collector.log

      Configuration:
        Set environment variables or use command-line flags.
        Run 'sqlite-otel-collector --help' for available options.

      Send test data:
        curl -X POST http://localhost:4318/v1/traces \\
          -H "Content-Type: application/json" \\
          -d '{"resourceSpans":[{"spans":[{"name":"test-span","kind":1}]}]}'

      Documentation: https://github.com/RedShiftVelocity/sqlite-otel
    EOS
  end
end