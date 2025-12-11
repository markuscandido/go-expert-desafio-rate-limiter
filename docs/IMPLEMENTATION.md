# Implementation Details

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│            RateLimiterMiddleware (middleware/)              │
│  - Extracts IP address from request                         │
│  - Extracts API_KEY token from header                       │
│  - Calls rate limiter logic                                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│            RateLimiter (limiter/)                           │
│  - Checks token limit (takes precedence)                    │
│  - Falls back to IP limit                                   │
│  - Returns allowed/blocked status                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│       Storage Strategy Interface (storage/)                 │
│  - CheckAndIncrement: Increment counter if allowed          │
│  - IsBlocked: Check if identifier is blocked               │
│  - Block: Block identifier for duration                     │
│  - GetData: Retrieve current limiter data                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│         Redis Implementation (storage/redis.go)             │
│  - Stores counters in Redis                                 │
│  - Uses TTL for automatic cleanup                           │
│  - Atomic operations for thread safety                      │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                      Redis Database                         │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Configuration (config/)

**Purpose**: Load and validate configuration from environment variables.

**Files**:
- `config.go`: Defines the `RateLimiterConfig` struct with default values
- `loader.go`: Loads configuration from `.env` file and environment variables

**Features**:
- Environment variable override system
- Sensible defaults
- Support for both `.env` file and direct environment variables

### 2. Storage Layer (storage/)

**Purpose**: Abstract storage operations behind an interface for flexibility.

**Files**:
- `strategy.go`: Defines the `Strategy` interface
- `redis.go`: Redis implementation

**Strategy Interface Methods**:

```go
type Strategy interface {
    CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (allowed bool, err error)
    IsBlocked(ctx context.Context, key string) (blocked bool, err error)
    Block(ctx context.Context, key string, durationSeconds int) error
    Reset(ctx context.Context, key string) error
    GetData(ctx context.Context, key string) (*LimiterData, error)
    Close() error
}
```

**Redis Data Format**:

Counters are stored as JSON:
```json
{
  "count": 5,
  "expires_at": "2024-12-11T10:30:45Z",
  "is_blocked": false
}
```

Keys use format:
- IP limiter: `ip:<ip_address>`
- Token limiter: `token:<api_token>`
- Blocked status: `<identifier>:blocked`

### 3. Rate Limiter Logic (limiter/)

**Purpose**: Implement the core rate limiting algorithm.

**Algorithm**:

```
For each request:
1. Extract IP and Token from request
2. If token limit enabled and token provided:
   a. Check if token is blocked
   b. Increment and check token counter
   c. Block if exceeded
3. Else if IP limit enabled:
   a. Check if IP is blocked
   b. Increment and check IP counter
   c. Block if exceeded
4. Allow request if not blocked
```

**Key Features**:
- Token precedence over IP limit
- Per-second sliding window
- Atomic operations via Redis
- Independent counters per identifier

### 4. Middleware (middleware/)

**Purpose**: HTTP middleware for easy integration with any web server.

**Handler Flow**:
```
1. Extract client IP (with proxy support)
2. Extract API token from API_KEY header
3. Call rate limiter
4. If blocked: return 429 Too Many Requests
5. If allowed: pass to next handler
```

**IP Extraction Priority**:
1. `X-Forwarded-For` header (for proxies)
2. `X-Real-IP` header
3. `RemoteAddr` from request

## Rate Limiting Algorithm

### Sliding Window Counter

The rate limiter uses a fixed window approach with the following logic:

```
1. Request arrives at time T
2. Check if we have data for this identifier
   - If not, create new counter with window = now + 1 second
3. If current time > window expiration:
   - Reset counter to 0
   - Set new window = now + 1 second
4. Increment counter
5. If counter > max_requests:
   - Block identifier for block_duration seconds
   - Return false (blocked)
6. Return true (allowed)
```

### Example Execution

```
Time: 00:00:00 - Request from IP 192.168.1.1
  Counter: 1, Window: 00:00:01

Time: 00:00:00.2 - Request from IP 192.168.1.1
  Counter: 2, Window: 00:00:01

Time: 00:00:00.4 - Request from IP 192.168.1.1
  Counter: 3, Window: 00:00:01

Time: 00:00:00.6 - Request from IP 192.168.1.1
  Counter: 4, Window: 00:00:01

Time: 00:00:00.8 - Request from IP 192.168.1.1
  Counter: 5, Window: 00:00:01 (Limit reached)

Time: 00:00:01.2 - Request from IP 192.168.1.1
  Blocked: IP blocked until 00:01:01

Time: 00:01:02 - Request from IP 192.168.1.1
  Counter: 1, Window: 00:01:03 (Block expired, new window)
```

## Redis Operations

### CheckAndIncrement Operation

```go
1. Check if blocked (GET ip:192.168.1.1:blocked)
2. Get current data (GET ip:192.168.1.1)
3. If expired or new:
   - Create new data with current timestamp + window
4. Increment counter
5. Store (SET ip:192.168.1.1 <data> EX <ttl>)
6. Return counter <= maxRequests
```

**Atomicity**: While individual operations aren't atomic, the short duration and idempotent nature make race conditions unlikely and non-critical.

### Block Operation

```go
SET <identifier>:blocked "true" EX <duration>
```

## Configuration Priority

1. Environment variables (highest priority)
2. `.env` file
3. Default values (lowest priority)

## Thread Safety

- **Go**: Safe due to goroutine-per-request model
- **Redis**: Atomic operations ensure consistency
- **No shared mutable state**: Each identifier tracked independently

## Error Handling

### Storage Errors
- Connection errors logged and propagated
- Failed requests return error instead of allowing/blocking
- Caller should handle gracefully (e.g., log and continue)

### Configuration Errors
- Invalid numbers logged as warnings
- Defaults used as fallback
- Application continues running

## Performance Characteristics

### Time Complexity
- CheckAndIncrement: O(1) with Redis
- IsBlocked: O(1) with Redis
- Block: O(1) with Redis

### Space Complexity
- O(n) where n = number of unique IPs/tokens
- Automatic cleanup via Redis TTL

### Latency
- Single request: 1-2 Redis operations
- Average latency: <1ms (depends on Redis latency)
- 99th percentile: <5ms

## Testing Strategy

### Unit Tests
- Mock storage implementation
- Test rate limiting logic independently
- Test configuration loading
- Test middleware separately

### Integration Tests
- Mock storage with time windows
- Test middleware + limiter together
- Test HTTP response codes
- Test IP/token extraction

### Manual Tests
- Load testing with concurrent requests
- Test with real Redis
- Test configuration loading
- Test block duration

## Future Enhancements

1. **Multiple storage backends**
   - Memcached
   - DynamoDB
   - PostgreSQL

2. **Advanced features**
   - Distributed rate limiting with consensus
   - Rate limit quota reset endpoints
   - Real-time monitoring dashboard
   - Metrics export (Prometheus)

3. **Optimizations**
   - Local cache layer for hot keys
   - Batch operations
   - Connection pooling

4. **Security**
   - Rate limit on rate limiter itself
   - DDoS protection
   - Request signature verification
