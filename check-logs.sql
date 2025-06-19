-- Check resources
SELECT 'Resources:' as section;
SELECT id, attributes, schema_url FROM resources;

-- Check scopes
SELECT '' as blank;
SELECT 'Instrumentation Scopes:' as section;
SELECT id, name, version, attributes FROM instrumentation_scopes;

-- Check log records
SELECT '' as blank;
SELECT 'Log Records:' as section;
SELECT 
    l.id,
    datetime(l.time_unix_nano/1000000000, 'unixepoch') as time,
    datetime(l.observed_time_unix_nano/1000000000, 'unixepoch') as observed_time,
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
JOIN instrumentation_scopes s ON l.scope_id = s.id;

-- Summary
SELECT '' as blank;
SELECT 'Summary:' as section;
SELECT 'Total log records: ' || COUNT(*) as summary FROM log_records;