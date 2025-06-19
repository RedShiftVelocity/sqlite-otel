# Resource & Scope Deduplication

This document describes the deduplication strategy for resources and instrumentation scopes.

## Overview

To prevent duplicate entries in the database, we implement a "find-or-create" pattern for:
- Resources (based on attributes and schema_url)
- Instrumentation Scopes (based on name, version, attributes, and schema_url)

## Implementation

### Resources
Before inserting a new resource, check if an identical one already exists:
```sql
SELECT id FROM resources 
WHERE attributes = ? AND schema_url = ?
```

### Instrumentation Scopes
Before inserting a new scope, check if an identical one already exists:
```sql
SELECT id FROM instrumentation_scopes 
WHERE name = ? AND version = ? AND attributes = ? AND schema_url = ?
```

## Benefits
- Reduces database size
- Improves query performance
- Maintains data integrity
- Enables efficient lookups

## Required Indexes
- `CREATE UNIQUE INDEX idx_resources_lookup ON resources(attributes, schema_url)`
- `CREATE UNIQUE INDEX idx_scopes_lookup ON instrumentation_scopes(name, version, attributes, schema_url)`