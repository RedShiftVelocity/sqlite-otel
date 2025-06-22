class SqliteOtelCollector < Formula
  desc "Lightweight OpenTelemetry collector with SQLite storage"
  homepage "https://github.com/RedShiftVelocity/sqlite-otel"
  version "0.8.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/RedShiftVelocity/sqlite-otel/releases/download/v0.8.0/sqlite-otel-darwin-amd64"
      sha256 "248838a90871f2baa46e9d58a22df0cbde1e1746819ac5e7904fced42cb98746"

      def install
        bin.install "sqlite-otel-darwin-amd64" => "sqlite-otel-collector"
      end
    end

    if Hardware::CPU.arm?
      url "https://github.com/RedShiftVelocity/sqlite-otel/releases/download/v0.8.0/sqlite-otel-darwin-arm64"
      sha256 "0dd76ddc6e69477bdb6b9f3b9c3d8e4843bcde35ac96f5b754b6092791ae0822"

      def install
        bin.install "sqlite-otel-darwin-arm64" => "sqlite-otel-collector"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/RedShiftVelocity/sqlite-otel/releases/download/v0.8.0/sqlite-otel-linux-amd64"
      sha256 "b4c04b74755e00ac6935ae96dd1c5195579c993c490a7c5ebf1e7d89b48da9f3"

      def install
        bin.install "sqlite-otel-linux-amd64" => "sqlite-otel-collector"
      end
    end

    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/RedShiftVelocity/sqlite-otel/releases/download/v0.8.0/sqlite-otel-linux-arm64"
      sha256 "e8e718977d8d381f62af725bf956ebab7bbe05edb4cc2517260487d9cdcb1f86"

      def install
        bin.install "sqlite-otel-linux-arm64" => "sqlite-otel-collector"
      end
    end

    if Hardware::CPU.arm? && !Hardware::CPU.is_64_bit?
      url "https://github.com/RedShiftVelocity/sqlite-otel/releases/download/v0.8.0/sqlite-otel-linux-arm"
      sha256 "5a47edcdb6d6b4a8a65e95e3081086a841836d70a17556fb960351e4cd6482e0"

      def install
        bin.install "sqlite-otel-linux-arm" => "sqlite-otel-collector"
      end
    end
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
    assert_match "sqlite-otel-collector v#{version}", output

    # Test that help command works
    output = shell_output("#{bin}/sqlite-otel-collector --help 2>&1")
    assert_match "Port to listen on", output
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