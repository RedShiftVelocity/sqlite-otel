package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var dbPath string
	flag.StringVar(&dbPath, "db-path", "./otel-data.db", "Path to SQLite database")
	flag.Parse()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("\n=== TRACES (SPANS) ===")
	printTraces(db)

	fmt.Println("\n=== METRICS ===")
	printMetrics(db)

	fmt.Println("\n=== LOGS ===")
	printLogs(db)
}

func printTraces(db *sql.DB) {
	query := `
		SELECT 
			s.trace_id, s.span_id, s.parent_span_id, s.name,
			s.start_time_unix_nano, s.end_time_unix_nano,
			s.attributes, s.status_code, s.status_message,
			r.attributes as resource_attrs,
			i.name as scope_name, i.version as scope_version
		FROM spans s
		LEFT JOIN resources r ON s.resource_id = r.id
		LEFT JOIN instrumentation_scopes i ON s.scope_id = i.id
		ORDER BY s.start_time_unix_nano DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying traces: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var traceID, spanID, parentSpanID, name sql.NullString
		var startTime, endTime sql.NullInt64
		var attributes, resourceAttrs sql.NullString
		var statusCode sql.NullInt64
		var statusMessage, scopeName, scopeVersion sql.NullString

		err := rows.Scan(&traceID, &spanID, &parentSpanID, &name,
			&startTime, &endTime, &attributes, &statusCode, &statusMessage,
			&resourceAttrs, &scopeName, &scopeVersion)
		if err != nil {
			log.Printf("Error scanning trace row: %v", err)
			continue
		}

		count++
		fmt.Printf("\nTrace %d:\n", count)
		fmt.Printf("  Trace ID: %s\n", traceID.String)
		fmt.Printf("  Span ID: %s\n", spanID.String)
		if parentSpanID.Valid {
			fmt.Printf("  Parent Span ID: %s\n", parentSpanID.String)
		}
		fmt.Printf("  Name: %s\n", name.String)
		
		if startTime.Valid {
			fmt.Printf("  Start Time: %s\n", formatNanoTime(startTime.Int64))
		}
		if endTime.Valid {
			fmt.Printf("  End Time: %s\n", formatNanoTime(endTime.Int64))
			if startTime.Valid {
				duration := time.Duration(endTime.Int64 - startTime.Int64)
				fmt.Printf("  Duration: %v\n", duration)
			}
		}

		if attributes.Valid && attributes.String != "" {
			fmt.Printf("  Attributes: %s\n", prettyJSON(attributes.String))
		}
		if resourceAttrs.Valid && resourceAttrs.String != "" {
			fmt.Printf("  Resource: %s\n", prettyJSON(resourceAttrs.String))
		}
		fmt.Printf("  Scope: %s@%s\n", scopeName.String, scopeVersion.String)
	}

	if count == 0 {
		fmt.Println("No traces found")
	}
}

func printMetrics(db *sql.DB) {
	query := `
		SELECT 
			m.name, m.description, m.unit, m.metric_type,
			d.time_unix_nano, d.value_double, d.value_int, d.attributes,
			r.attributes as resource_attrs,
			i.name as scope_name, i.version as scope_version
		FROM metrics m
		JOIN metric_data_points d ON m.id = d.metric_id
		LEFT JOIN resources r ON m.resource_id = r.id
		LEFT JOIN instrumentation_scopes i ON m.scope_id = i.id
		ORDER BY d.time_unix_nano DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying metrics: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name, description, unit, metricType sql.NullString
		var timestamp sql.NullInt64
		var valueDouble sql.NullFloat64
		var valueInt sql.NullInt64
		var attributes, resourceAttrs sql.NullString
		var scopeName, scopeVersion sql.NullString

		err := rows.Scan(&name, &description, &unit, &metricType,
			&timestamp, &valueDouble, &valueInt, &attributes,
			&resourceAttrs, &scopeName, &scopeVersion)
		if err != nil {
			log.Printf("Error scanning metric row: %v", err)
			continue
		}

		count++
		fmt.Printf("\nMetric %d:\n", count)
		fmt.Printf("  Name: %s\n", name.String)
		if description.Valid && description.String != "" {
			fmt.Printf("  Description: %s\n", description.String)
		}
		if unit.Valid && unit.String != "" {
			fmt.Printf("  Unit: %s\n", unit.String)
		}
		fmt.Printf("  Type: %s\n", metricType.String)
		
		if timestamp.Valid {
			fmt.Printf("  Time: %s\n", formatNanoTime(timestamp.Int64))
		}
		
		if valueDouble.Valid {
			fmt.Printf("  Value: %f\n", valueDouble.Float64)
		} else if valueInt.Valid {
			fmt.Printf("  Value: %d\n", valueInt.Int64)
		}

		if attributes.Valid && attributes.String != "" {
			fmt.Printf("  Attributes: %s\n", prettyJSON(attributes.String))
		}
		if resourceAttrs.Valid && resourceAttrs.String != "" {
			fmt.Printf("  Resource: %s\n", prettyJSON(resourceAttrs.String))
		}
		fmt.Printf("  Scope: %s@%s\n", scopeName.String, scopeVersion.String)
	}

	if count == 0 {
		fmt.Println("No metrics found")
	}
}

func printLogs(db *sql.DB) {
	query := `
		SELECT 
			l.time_unix_nano, l.observed_time_unix_nano,
			l.severity_number, l.severity_text, l.body,
			l.attributes, l.trace_id, l.span_id,
			r.attributes as resource_attrs,
			i.name as scope_name, i.version as scope_version
		FROM log_records l
		LEFT JOIN resources r ON l.resource_id = r.id
		LEFT JOIN instrumentation_scopes i ON l.scope_id = i.id
		ORDER BY l.time_unix_nano DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying logs: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var timestamp, observedTime sql.NullInt64
		var severityNum sql.NullInt64
		var severityText, body, attributes sql.NullString
		var traceID, spanID sql.NullString
		var resourceAttrs sql.NullString
		var scopeName, scopeVersion sql.NullString

		err := rows.Scan(&timestamp, &observedTime, &severityNum, &severityText,
			&body, &attributes, &traceID, &spanID,
			&resourceAttrs, &scopeName, &scopeVersion)
		if err != nil {
			log.Printf("Error scanning log row: %v", err)
			continue
		}

		count++
		fmt.Printf("\nLog %d:\n", count)
		
		if timestamp.Valid {
			fmt.Printf("  Time: %s\n", formatNanoTime(timestamp.Int64))
		}
		if observedTime.Valid {
			fmt.Printf("  Observed Time: %s\n", formatNanoTime(observedTime.Int64))
		}
		
		if severityNum.Valid {
			fmt.Printf("  Severity: %d", severityNum.Int64)
			if severityText.Valid && severityText.String != "" {
				fmt.Printf(" (%s)", severityText.String)
			}
			fmt.Println()
		}
		
		if body.Valid && body.String != "" {
			fmt.Printf("  Body: %s\n", body.String)
		}
		
		if traceID.Valid && traceID.String != "" {
			fmt.Printf("  Trace ID: %s\n", traceID.String)
		}
		if spanID.Valid && spanID.String != "" {
			fmt.Printf("  Span ID: %s\n", spanID.String)
		}

		if attributes.Valid && attributes.String != "" {
			fmt.Printf("  Attributes: %s\n", prettyJSON(attributes.String))
		}
		if resourceAttrs.Valid && resourceAttrs.String != "" {
			fmt.Printf("  Resource: %s\n", prettyJSON(resourceAttrs.String))
		}
		fmt.Printf("  Scope: %s@%s\n", scopeName.String, scopeVersion.String)
	}

	if count == 0 {
		fmt.Println("No logs found")
	}
}

func formatNanoTime(nanos int64) string {
	t := time.Unix(0, nanos)
	return t.Format("2006-01-02 15:04:05.000000000 MST")
}

func prettyJSON(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr
	}
	pretty, err := json.MarshalIndent(data, "    ", "  ")
	if err != nil {
		return jsonStr
	}
	return "\n    " + string(pretty)
}