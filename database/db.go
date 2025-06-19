package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create all required tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func createTables() error {
	// Resource table (shared across all telemetry types)
	resourcesTable := `
	CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		attributes TEXT NOT NULL,
		schema_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Instrumentation Scope table (shared across all telemetry types)
	scopesTable := `
	CREATE TABLE IF NOT EXISTS instrumentation_scopes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		version TEXT,
		attributes TEXT,
		schema_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Main spans table
	spansTable := `
	CREATE TABLE IF NOT EXISTS spans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		trace_id TEXT NOT NULL,
		span_id TEXT NOT NULL,
		parent_span_id TEXT,
		trace_state TEXT,
		name TEXT NOT NULL,
		kind INTEGER NOT NULL,
		start_time_unix_nano INTEGER NOT NULL,
		end_time_unix_nano INTEGER NOT NULL,
		attributes TEXT,
		dropped_attributes_count INTEGER DEFAULT 0,
		status_code INTEGER DEFAULT 0,
		status_message TEXT,
		flags INTEGER DEFAULT 0,
		resource_id INTEGER,
		scope_id INTEGER,
		ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resource_id) REFERENCES resources(id),
		FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
	);`

	// Span events table
	spanEventsTable := `
	CREATE TABLE IF NOT EXISTS span_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		span_id INTEGER NOT NULL,
		time_unix_nano INTEGER NOT NULL,
		name TEXT NOT NULL,
		attributes TEXT,
		dropped_attributes_count INTEGER DEFAULT 0,
		FOREIGN KEY (span_id) REFERENCES spans(id) ON DELETE CASCADE
	);`

	// Span links table
	spanLinksTable := `
	CREATE TABLE IF NOT EXISTS span_links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		span_id INTEGER NOT NULL,
		trace_id TEXT NOT NULL,
		span_id_linked TEXT NOT NULL,
		trace_state TEXT,
		attributes TEXT,
		dropped_attributes_count INTEGER DEFAULT 0,
		flags INTEGER DEFAULT 0,
		FOREIGN KEY (span_id) REFERENCES spans(id) ON DELETE CASCADE
	);`

	// Main metrics table
	metricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		unit TEXT,
		type TEXT NOT NULL,
		resource_id INTEGER,
		scope_id INTEGER,
		ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resource_id) REFERENCES resources(id),
		FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
	);`

	// Metric data points table
	metricDataPointsTable := `
	CREATE TABLE IF NOT EXISTS metric_data_points (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		metric_id INTEGER NOT NULL,
		attributes TEXT,
		start_time_unix_nano INTEGER,
		time_unix_nano INTEGER NOT NULL,
		flags INTEGER DEFAULT 0,
		value_double REAL,
		value_int INTEGER,
		aggregation_temporality INTEGER,
		is_monotonic BOOLEAN,
		count INTEGER,
		sum_value REAL,
		min_value REAL,
		max_value REAL,
		bucket_counts TEXT,
		explicit_bounds TEXT,
		scale INTEGER,
		zero_count INTEGER,
		positive_offset INTEGER,
		positive_bucket_counts TEXT,
		negative_offset INTEGER,
		negative_bucket_counts TEXT,
		quantile_values TEXT,
		FOREIGN KEY (metric_id) REFERENCES metrics(id) ON DELETE CASCADE
	);`

	// Main log records table
	logRecordsTable := `
	CREATE TABLE IF NOT EXISTS log_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		time_unix_nano INTEGER,
		observed_time_unix_nano INTEGER NOT NULL,
		severity_number INTEGER,
		severity_text TEXT,
		body TEXT,
		attributes TEXT,
		trace_id TEXT,
		span_id TEXT,
		trace_flags INTEGER DEFAULT 0,
		resource_id INTEGER,
		scope_id INTEGER,
		ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resource_id) REFERENCES resources(id),
		FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
	);`

	// Execute all table creation queries
	tables := []string{
		resourcesTable,
		scopesTable,
		spansTable,
		spanEventsTable,
		spanLinksTable,
		metricsTable,
		metricDataPointsTable,
		logRecordsTable,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes
	indexes := []string{
		// Indexes for spans
		"CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id)",
		"CREATE INDEX IF NOT EXISTS idx_spans_span_id ON spans(span_id)",
		"CREATE INDEX IF NOT EXISTS idx_spans_parent_span_id ON spans(parent_span_id)",
		"CREATE INDEX IF NOT EXISTS idx_spans_start_time ON spans(start_time_unix_nano)",
		"CREATE INDEX IF NOT EXISTS idx_spans_name ON spans(name)",
		"CREATE INDEX IF NOT EXISTS idx_span_events_span_id ON span_events(span_id)",
		"CREATE INDEX IF NOT EXISTS idx_span_links_span_id ON span_links(span_id)",
		// Indexes for metrics
		"CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name)",
		"CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type)",
		"CREATE INDEX IF NOT EXISTS idx_metric_data_points_metric_id ON metric_data_points(metric_id)",
		"CREATE INDEX IF NOT EXISTS idx_metric_data_points_time ON metric_data_points(time_unix_nano)",
		// Indexes for logs
		"CREATE INDEX IF NOT EXISTS idx_log_records_time ON log_records(time_unix_nano)",
		"CREATE INDEX IF NOT EXISTS idx_log_records_observed_time ON log_records(observed_time_unix_nano)",
		"CREATE INDEX IF NOT EXISTS idx_log_records_severity ON log_records(severity_number)",
		"CREATE INDEX IF NOT EXISTS idx_log_records_trace_id ON log_records(trace_id)",
		"CREATE INDEX IF NOT EXISTS idx_log_records_span_id ON log_records(span_id)",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}