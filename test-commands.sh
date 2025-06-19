#!/bin/bash

# Sample OTLP test commands
# Run the collector with either:
#   ./sqlite-otel-collector              # Random ephemeral port
#   ./sqlite-otel-collector -port 4318   # Specific port (4318 is OTLP/HTTP default)
# Note the port number it prints, then replace PORT below with that number

PORT=4319

echo "Testing OTLP endpoints..."
echo "Replace PORT=$PORT with the actual port number from the collector output"
echo ""

# Test traces endpoint with sample JSON data
echo "1. Testing /v1/traces endpoint:"
echo "curl -X POST http://localhost:$PORT/v1/traces \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"resourceSpans\":[{\"resource\":{\"attributes\":[{\"key\":\"service.name\",\"value\":{\"stringValue\":\"test-service\"}}]},\"scopeSpans\":[{\"spans\":[{\"traceId\":\"5B8EFFF798038103D269B633813FC60C\",\"spanId\":\"EEE19B7EC3C1B174\",\"name\":\"test-span\",\"startTimeUnixNano\":\"1544712660000000000\",\"endTimeUnixNano\":\"1544712661000000000\"}]}]}]}'"
echo ""

# Test metrics endpoint with sample JSON data
echo "2. Testing /v1/metrics endpoint:"
echo "curl -X POST http://localhost:$PORT/v1/metrics \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"resourceMetrics\":[{\"resource\":{\"attributes\":[{\"key\":\"service.name\",\"value\":{\"stringValue\":\"test-service\"}}]},\"scopeMetrics\":[{\"metrics\":[{\"name\":\"test.counter\",\"unit\":\"1\",\"sum\":{\"dataPoints\":[{\"asInt\":\"100\",\"timeUnixNano\":\"1544712660000000000\"}]}}]}]}]}'"
echo ""

# Test logs endpoint with sample JSON data
echo "3. Testing /v1/logs endpoint:"
echo "curl -X POST http://localhost:$PORT/v1/logs \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"resourceLogs\":[{\"resource\":{\"attributes\":[{\"key\":\"service.name\",\"value\":{\"stringValue\":\"test-service\"}}]},\"scopeLogs\":[{\"logRecords\":[{\"timeUnixNano\":\"1544712660000000000\",\"severityText\":\"INFO\",\"body\":{\"stringValue\":\"Test log message\"}}]}]}]}'"
echo ""

# Simple test with minimal data
echo "4. Simple test (traces):"
echo "curl -X POST http://localhost:$PORT/v1/traces \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"test\":\"data\"}'"
echo ""

# Test with wrong Content-Type (should fail)
echo "5. Test with wrong Content-Type (should return 415 Unsupported Media Type):"
echo "curl -X POST http://localhost:$PORT/v1/traces \\"
echo "  -H 'Content-Type: application/protobuf' \\"
echo "  -d '{\"test\":\"data\"}'"