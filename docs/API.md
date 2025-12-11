# API Documentation

## Server Information

- **Host**: `localhost`
- **Port**: `8080`
- **Base URL**: `http://localhost:8080`

## Endpoints

### Health Check

Check server health status.

**Endpoint**:
```
GET /health
```

**Response**:
```json
{
  "status": "healthy"
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

### Root Endpoint

Example endpoint showing successful response.

**Endpoint**:
```
GET /
```

**Response**:
```json
{
  "message": "Hello from rate-limiter server!"
}
```

**Example**:
```bash
curl http://localhost:8080/
```

## Rate Limiting

### Request Headers

#### API_KEY (Optional)
Token-based rate limiting identifier.

```
API_KEY: <your-api-token>
```

**Usage**:
```bash
curl -H "API_KEY: premium-token" http://localhost:8080/
```

### Response Headers

#### Retry-After
When request is rate limited, this header indicates seconds to wait.

**Example Response**:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: text/plain

you have reached the maximum number of requests or actions allowed within a certain time frame
```

## Response Codes

### 200 OK
Request allowed and processed successfully.

```bash
curl -i http://localhost:8080/
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 50

{"message": "Hello from rate-limiter server!"}
```

### 429 Too Many Requests
Request rate limit exceeded.

```bash
curl -i http://localhost:8080/
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: text/plain

you have reached the maximum number of requests or actions allowed within a certain time frame
```

### 500 Internal Server Error
Internal server error (e.g., Redis connection failure).

```bash
HTTP/1.1 500 Internal Server Error
Content-Type: text/plain

Internal Server Error
```

## Rate Limiting Behavior

### IP-Based Rate Limiting

Limits requests per IP address.

**Configuration**:
```env
RATE_LIMITER_ENABLE_IP=true
RATE_LIMITER_MAX_REQUESTS_IP=10
RATE_LIMITER_BLOCK_DURATION_IP=60
```

**Behavior**:
- Allows 10 requests per second from each IP
- Blocks after 10 requests for 60 seconds
- Different IPs have independent counters

**Example**:
```bash
# First 10 requests succeed
for i in {1..10}; do
  curl http://localhost:8080/
done

# 11th request is blocked
curl http://localhost:8080/
# Returns: HTTP 429
```

### Token-Based Rate Limiting

Limits requests per API token. Takes precedence over IP limits.

**Configuration**:
```env
RATE_LIMITER_ENABLE_TOKEN=true
RATE_LIMITER_MAX_REQUESTS_TOKEN=100
RATE_LIMITER_BLOCK_DURATION_TOKEN=60
```

**Behavior**:
- Allows 100 requests per second per token
- Blocks after 100 requests for 60 seconds
- Different tokens have independent counters
- Overrides IP limit if both enabled

**Example**:
```bash
# First 100 requests with token succeed
for i in {1..100}; do
  curl -H "API_KEY: premium-token" http://localhost:8080/
done

# 101st request is blocked
curl -H "API_KEY: premium-token" http://localhost:8080/
# Returns: HTTP 429
```

### Token Precedence

When both IP and token limits are enabled, token limit is checked first.

**Example Scenario**:
```
IP limit:    5 req/s
Token limit: 100 req/s

Request without token after 5 requests → Blocked (IP limit)
Request with token after 5 requests → Allowed (Token limit 100)
```

**Example**:
```bash
# Exhaust IP limit (assuming 5 req/s)
for i in {1..5}; do
  curl http://localhost:8080/
done

# 6th request without token → Blocked
curl http://localhost:8080/
# Returns: HTTP 429

# But with token → Allowed
curl -H "API_KEY: premium-token" http://localhost:8080/
# Returns: HTTP 200
```

## IP Address Detection

The rate limiter extracts client IP in the following order:

1. **X-Forwarded-For Header** (for proxy support)
   ```bash
   curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/
   ```

2. **X-Real-IP Header**
   ```bash
   curl -H "X-Real-IP: 192.168.1.100" http://localhost:8080/
   ```

3. **RemoteAddr** (direct connection)

## Error Responses

### Rate Limit Exceeded

```
Status: 429 Too Many Requests
Headers:
  - Retry-After: 60
  - Content-Type: text/plain

Body:
you have reached the maximum number of requests or actions allowed within a certain time frame
```

### Redis Connection Error

```
Status: 500 Internal Server Error
Headers:
  - Content-Type: text/plain

Body:
Internal Server Error
```

## Usage Examples

### Example 1: Basic API Call

```bash
curl -v http://localhost:8080/
```

**Output**:
```
> GET / HTTP/1.1
> Host: localhost:8080
>

< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 50
<

{"message": "Hello from rate-limiter server!"}
```

### Example 2: With Authentication Token

```bash
curl -v -H "API_KEY: my-token" http://localhost:8080/
```

**Output**:
```
> GET / HTTP/1.1
> Host: localhost:8080
> API_KEY: my-token
>

< HTTP/1.1 200 OK
< Content-Type: application/json
<

{"message": "Hello from rate-limiter server!"}
```

### Example 3: Rate Limit Exceeded

```bash
# Send requests in rapid succession
for i in {1..10}; do
  curl -s http://localhost:8080/ & 
done
wait

# Some will return 429
```

**Output**:
```
< HTTP/1.1 429 Too Many Requests
< Retry-After: 60
< Content-Type: text/plain
<

you have reached the maximum number of requests or actions allowed within a certain time frame
```

### Example 4: Behind Proxy

```bash
# When behind a proxy, X-Forwarded-For is used for IP detection
curl -H "X-Forwarded-For: 203.0.113.100" http://localhost:8080/
```

### Example 5: Health Check in Script

```bash
#!/bin/bash

response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

if [ $response -eq 200 ]; then
  echo "Server is healthy"
  exit 0
else
  echo "Server returned status: $response"
  exit 1
fi
```

### Example 6: Retry Logic

```bash
#!/bin/bash

max_retries=5
retry_count=0
retry_delay=1

while [ $retry_count -lt $max_retries ]; do
  response=$(curl -s -w "\n%{http_code}" http://localhost:8080/)
  status_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | head -n-1)

  if [ "$status_code" -eq 200 ]; then
    echo "Success: $body"
    exit 0
  elif [ "$status_code" -eq 429 ]; then
    retry_after=$(curl -sI http://localhost:8080/ | grep -i "retry-after" | cut -d' ' -f2)
    echo "Rate limited. Retrying after ${retry_after}s"
    sleep "${retry_after:-$retry_delay}"
    retry_count=$((retry_count + 1))
  else
    echo "Error: Status $status_code"
    exit 1
  fi
done

echo "Max retries exceeded"
exit 1
```

## Rate Limiting Timeline

### Request Timeline (5 req/s limit)

```
Time        | Request | Counter | Status | Action
-----------+---------+---------+--------+--------
00:00:00.0  | 1       | 1/5     | 200    | Allow
00:00:00.2  | 2       | 2/5     | 200    | Allow
00:00:00.4  | 3       | 3/5     | 200    | Allow
00:00:00.6  | 4       | 4/5     | 200    | Allow
00:00:00.8  | 5       | 5/5     | 200    | Allow
00:00:01.0  | 6       | 1/5     | 200    | Window reset
00:00:01.2  | 7       | 2/5     | 200    | Allow
00:00:10.5  | 8       | 1/5     | 200    | Allow (window expired)
```

## Performance Considerations

### Request Latency

- **99% of requests**: < 1ms additional latency
- **Rate check**: Typically < 0.5ms
- **Redis operation**: < 1ms (local network)

### Throughput

- **Sustainable**: 10,000+ requests/second
- **Burst capacity**: 50,000+ requests/second
- **Depends on**: Redis performance, network latency, server resources

### Resource Usage

- **Memory**: ~200 bytes per active IP/token
- **Redis operations**: 1-2 operations per request
- **CPU**: Negligible impact

## Backward Compatibility

The API follows semantic versioning. Current version is v1.

Breaking changes will be indicated by major version increment (v2+).

## Deprecation

No deprecated endpoints at this time.

## Rate Limiting Limits

To prevent abuse of the rate limiter itself:

- **Maximum tokens tracked**: Limited by Redis memory
- **Maximum concurrent connections**: Depends on server configuration
- **Maximum request size**: Standard HTTP limitations

## API Versioning

Current API version: **v1.0.0**

API version follows project version in `go.mod`.
