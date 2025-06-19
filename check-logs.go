package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "otel-collector.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check resources
	fmt.Println("=== Resources ===")
	rows, err := db.Query("SELECT id, attributes, schema_url FROM resources")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var attributes string
		var schemaURL sql.NullString
		if err := rows.Scan(&id, &attributes, &schemaURL); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Attributes: %s, Schema URL: %s\n", id, attributes, schemaURL.String)
	}

	// Check scopes
	fmt.Println("\n=== Instrumentation Scopes ===")
	rows, err = db.Query("SELECT id, name, version, attributes FROM instrumentation_scopes")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		var version sql.NullString
		var attributes string
		if err := rows.Scan(&id, &name, &version, &attributes); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Name: %s, Version: %s, Attributes: %s\n", id, name, version.String, attributes)
	}

	// Check log records
	fmt.Println("\n=== Log Records ===")
	rows, err = db.Query(`
		SELECT 
			l.id,
			l.time_unix_nano,
			l.observed_time_unix_nano,
			l.severity_number,
			l.severity_text,
			l.body,
			l.attributes,
			l.trace_id,
			l.span_id,
			l.flags,
			r.attributes as resource_attrs,
			s.name as scope_name
		FROM log_records l
		JOIN resources r ON l.resource_id = r.id
		JOIN instrumentation_scopes s ON l.scope_id = s.id
		ORDER BY l.id`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, timeUnix, observedTime, severityNumber, flags int64
		var severityText, body, attributes, traceID, spanID, resourceAttrs, scopeName string
		if err := rows.Scan(&id, &timeUnix, &observedTime, &severityNumber, &severityText,
			&body, &attributes, &traceID, &spanID, &flags, &resourceAttrs, &scopeName); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nLog Record ID: %d\n", id)
		if timeUnix > 0 {
			fmt.Printf("  Time: %s\n", time.Unix(0, timeUnix).Format(time.RFC3339Nano))
		}
		if observedTime > 0 {
			fmt.Printf("  Observed Time: %s\n", time.Unix(0, observedTime).Format(time.RFC3339Nano))
		}
		fmt.Printf("  Severity: %d (%s)\n", severityNumber, severityText)
		
		// Pretty print body
		var bodyObj interface{}
		if err := json.Unmarshal([]byte(body), &bodyObj); err == nil {
			bodyBytes, _ := json.MarshalIndent(bodyObj, "  ", "  ")
			fmt.Printf("  Body: %s\n", string(bodyBytes))
		} else {
			fmt.Printf("  Body: %s\n", body)
		}
		
		fmt.Printf("  Attributes: %s\n", attributes)
		if traceID != "" {
			fmt.Printf("  Trace ID: %s\n", traceID)
		}
		if spanID != "" {
			fmt.Printf("  Span ID: %s\n", spanID)
		}
		fmt.Printf("  Flags: %d\n", flags)
		fmt.Printf("  Resource: %s\n", resourceAttrs)
		fmt.Printf("  Scope: %s\n", scopeName)
		count++
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total log records: %d\n", count)
}