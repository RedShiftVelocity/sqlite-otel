# Data Integrity & Error Handling

## Overview

This document outlines the data integrity and error handling improvements needed for robust telemetry storage.

## Current Issues

1. **Silent failures**: Errors from `json.Marshal` are ignored
2. **Missing validation**: Required fields are not checked
3. **Type assertions**: The `ok` boolean from type assertions is not checked
4. **Data corruption**: Invalid data can be silently inserted as empty/null values

## Solutions

### 1. Error Checking for json.Marshal

```go
// Before (bad)
attributesJSON, _ := json.Marshal(attributes)

// After (good)
attributesJSON, err := json.Marshal(attributes)
if err != nil {
    return fmt.Errorf("failed to marshal attributes: %w", err)
}
```

### 2. Required Field Validation

```go
// Validate required fields
traceID, ok := span["traceId"].(string)
if !ok || traceID == "" {
    return fmt.Errorf("invalid span: traceId is required")
}
```

### 3. Type Assertion Checks

```go
// Check type assertions
name, ok := metric["name"].(string)
if !ok {
    return fmt.Errorf("invalid metric: name must be a string")
}
```

## Benefits

- No silent data corruption
- Clear error messages for debugging
- Data consistency guarantees
- Better observability of issues
- Compliance with OTLP specification