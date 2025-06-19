package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the SQLite database connection and creates tables
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// createTables creates all required tables
func createTables() error {
	tables := []string{
		// Resources table
		`CREATE TABLE IF NOT EXISTS resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			attributes TEXT,
			schema_url TEXT
		)`,

		// Instrumentation scopes table
		`CREATE TABLE IF NOT EXISTS instrumentation_scopes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			version TEXT,
			attributes TEXT,
			schema_url TEXT
		)`,

		// Spans table
		`CREATE TABLE IF NOT EXISTS spans (
			trace_id TEXT NOT NULL,
			span_id TEXT NOT NULL,
			trace_state TEXT,
			parent_span_id TEXT,
			name TEXT,
			kind INTEGER,
			start_time_unix_nano INTEGER,
			end_time_unix_nano INTEGER,
			attributes TEXT,
			events TEXT,
			links TEXT,
			status_code INTEGER,
			status_message TEXT,
			resource_id INTEGER,
			scope_id INTEGER,
			PRIMARY KEY (trace_id, span_id),
			FOREIGN KEY (resource_id) REFERENCES resources (id),
			FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes (id)
		)`,

		// Metrics table
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			unit TEXT,
			type TEXT NOT NULL,
			resource_id INTEGER,
			scope_id INTEGER,
			FOREIGN KEY (resource_id) REFERENCES resources (id),
			FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes (id)
		)`,

		// Metric data points table
		`CREATE TABLE IF NOT EXISTS metric_data_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_id INTEGER NOT NULL,
			attributes TEXT,
			start_time_unix_nano INTEGER,
			time_unix_nano INTEGER,
			value_double REAL,
			value_int INTEGER,
			exemplars TEXT,
			flags INTEGER,
			FOREIGN KEY (metric_id) REFERENCES metrics (id)
		)`,

		// Log records table
		`CREATE TABLE IF NOT EXISTS log_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time_unix_nano INTEGER,
			observed_time_unix_nano INTEGER,
			severity_number INTEGER,
			severity_text TEXT,
			body TEXT,
			attributes TEXT,
			trace_id TEXT,
			span_id TEXT,
			flags INTEGER,
			resource_id INTEGER,
			scope_id INTEGER,
			FOREIGN KEY (resource_id) REFERENCES resources (id),
			FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes (id)
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_spans_resource_id ON spans(resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_resource_id ON metrics(resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_log_records_trace_id ON log_records(trace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_log_records_resource_id ON log_records(resource_id)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}