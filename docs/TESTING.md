# Testing Guide

## Test Structure

The rate limiter includes comprehensive tests at multiple levels:

1. **Unit Tests**: Individual component testing
2. **Integration Tests**: Component interaction testing
3. **Manual Tests**: Real-world scenario testing

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Specific Test Package
```bash
go test -v ./limiter
go test -v ./middleware
go test -v ./storage
```

### Run Specific Test
```bash
go test -run TestIPRateLimiting ./limiter
```

### Run with Verbose Output
```bash
go test -v ./...
```

## Unit Tests

### Limiter Tests (`limiter/limiter_test.go`)

#### TestIPRateLimiting
Tests IP-based rate limiting:
- Allows requests up to limit
- Blocks excess requests
- Blocks the correct identifier
- Different IPs are independent

#### TestTokenRateLimiting
Tests token-based rate limiting:
- Allows requests up to token limit
- Blocks excess token requests
- Different tokens are independent

#### TestTokenPrecedenceOverIP
Tests that token limits override IP limits:
- Token limit applies instead of IP limit
- Higher limits for tokens work correctly

#### TestDisabledLimits
Tests behavior when all limits disabled:
- All requests allowed regardless of count

### Middleware Tests (`middleware/middleware_test.go`)

#### TestMiddlewareBlocksExceededRequests
Tests HTTP middleware behavior:
- Returns 200 for allowed requests
- Returns 429 for blocked requests
- Response body matches specification

#### TestMiddlewareWithToken
Tests middleware token extraction:
- Reads API_KEY header correctly
- Applies token-based limits
- Returns 429 when token limit exceeded

## Integration Tests (`integration_test.go`)

#### TestIPRateLimitingIntegration
End-to-end IP limiting test:
- Simulates HTTP requests
- Tests middleware + limiter interaction
- Verifies HTTP status codes

#### TestTokenRateLimitingIntegration
End-to-end token limiting test:
- HTTP requests with API_KEY header
- Verifies token limit enforcement

#### TestTokenPrecedenceIntegration
Tests token precedence with HTTP:
- IP blocked but token allowed
- Demonstrates precedence in action

#### TestDifferentIPsAreIndependent
Tests isolation between identifiers:
- Multiple IPs tracked independently
- Blocking one doesn't affect others

## Manual Testing

### Setup
```bash
# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Build the application
go build -o rate-limiter .

# Run the application
./rate-limiter
```

### Test 1: Basic IP Limiting

```bash
# Config: 5 requests/sec from same IP
export RATE_LIMITER_MAX_REQUESTS_IP=5

# Test script
for i in {1..10}; do
  echo "Request $i:"
  curl -w "HTTP %{http_code}\n" http://localhost:8080/
  sleep 0.1
done
```

**Expected**:
- Requests 1-5: HTTP 200
- Requests 6-10: HTTP 429

### Test 2: Token Limiting

```bash
# Config: 3 requests/sec for token
export RATE_LIMITER_MAX_REQUESTS_TOKEN=3

# Test script
for i in {1..5}; do
  echo "Request $i with token:"
  curl -w "HTTP %{http_code}\n" \
    -H "API_KEY: test-token" \
    http://localhost:8080/
  sleep 0.1
done
```

**Expected**:
- Requests 1-3: HTTP 200
- Requests 4-5: HTTP 429

### Test 3: Block Duration

```bash
# Make request to trigger block (assuming IP limit = 5)
for i in {1..6}; do
  curl -w "HTTP %{http_code}\n" http://localhost:8080/
done

# Wait 5 seconds (less than 60 second block)
sleep 5

# Should still be blocked
curl -w "HTTP %{http_code}\n" http://localhost:8080/
# Returns: HTTP 429

# Wait 60 more seconds (total > 60)
sleep 60

# Should be allowed now
curl -w "HTTP %{http_code}\n" http://localhost:8080/
# Returns: HTTP 200
```

### Test 4: Different IPs

```bash
# IP 1: 192.168.1.1 - should be limited
for i in {1..6}; do
  curl -w "HTTP %{http_code}\n" \
    --header "X-Forwarded-For: 192.168.1.1" \
    http://localhost:8080/
done

# IP 2: 192.168.1.2 - should be independent
for i in {1..6}; do
  curl -w "HTTP %{http_code}\n" \
    --header "X-Forwarded-For: 192.168.1.2" \
    http://localhost:8080/
done
```

### Test 5: Load Testing

```bash
# Using Apache Bench
ab -n 100 -c 10 http://localhost:8080/

# Using wrk (if installed)
wrk -t4 -c100 -d10s http://localhost:8080/
```

### Test 6: Configuration Reload

```bash
# Create .env file
cat > .env << EOF
RATE_LIMITER_MAX_REQUESTS_IP=2
RATE_LIMITER_BLOCK_DURATION_IP=30
EOF

# Rebuild and restart
go build -o rate-limiter .
./rate-limiter

# Test (should allow only 2 requests)
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/  # Should be blocked
```

### Test 7: Redis Connection Failure

```bash
# Stop Redis
docker stop <redis-container>

# Try to start rate limiter
./rate-limiter
# Should fail with connection error

# Restart Redis
docker start <redis-container>

# Restart rate limiter
./rate-limiter
# Should connect successfully
```

## Test Coverage

View coverage with:
```bash
go test -cover ./...
```

Generate HTML coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Target coverage: >80% for production code

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test -v -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v2
```

## Debugging Tests

### Verbose Output
```bash
go test -v ./limiter
```

### Run Single Test with Debug
```bash
go test -run TestIPRateLimiting -v ./limiter -count=1
```

### Show All Logs
```bash
go test -v ./... 2>&1 | tee test.log
```

## Performance Testing

### Benchmark
```bash
go test -bench=. -benchmem ./limiter
```

### Stress Test
```bash
# Create load test script
cat > stress_test.sh << 'EOF'
#!/bin/bash
for i in {1..10000}; do
  curl -s http://localhost:8080/ > /dev/null
done
EOF

chmod +x stress_test.sh
time ./stress_test.sh
```

## Test Data

### Sample Redis Data
```bash
redis-cli

# View all keys
KEYS *

# View rate limiter data
GET "ip:192.168.1.1"
GET "token:abc123:blocked"

# View TTL
TTL "ip:192.168.1.1"
```

### Test Fixtures

Mock storage is used for unit tests to avoid Redis dependency:

```go
type MockStrategy struct {
    data    map[string]*storage.LimiterData
    blocked map[string]bool
}
```

## Troubleshooting Failed Tests

### Test Hangs
- Redis connection timeout
- Solution: Start Redis or increase timeout

### Flaky Tests
- Race conditions on shared mock data
- Solution: Reset mock between tests, use context with timeout

### Import Errors
- Missing dependencies
- Solution: `go mod download && go mod tidy`

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Reset mock storage between tests
3. **Assertions**: Clear error messages on failure
4. **Coverage**: Aim for >80% code coverage
5. **Performance**: Tests should run in <10 seconds total
6. **Documentation**: Document complex test scenarios

## Running Tests in Docker

```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go test -v ./...
```

```bash
docker build -f Dockerfile.test -t rate-limiter-test .
docker run rate-limiter-test
```
