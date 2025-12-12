# Guia de Uso e Configuração do Logger

## Instalação

O logger já está integrado no projeto. Basta importar:

```go
import "github.com/markuscandido/go-expert-desafio-rate-limiter/pkg/logger"
```

## Níveis de Log

O projeto suporta os seguintes níveis de log (em ordem de severidade):

1. **DEBUG** - Informações detalhadas para diagnóstico
2. **INFO** - Eventos normais e significativos
3. **WARN** - Eventos incomuns que devem ser investigados
4. **ERROR** - Erros que afetam funcionalidades
5. **FATAL** - Erros críticos que interrompem a execução

## Uso Básico

### Info
```go
logger.Info("User logged in", "userID", user.ID, "ip", r.RemoteAddr)
```

### Error
```go
logger.Error("Database connection failed", "error", err, "retries", 3)
```

### Warn
```go
logger.Warn("Rate limit approaching", "current", 95, "limit", 100)
```

### Debug
```go
logger.Debug("Processing request", "method", r.Method, "path", r.URL.Path)
```

### Fatal
```go
logger.Fatal("Cannot start server", "port", port, "error", err)
```

## Padrões de Logging

### 1. Sempre incluir contexto relevante
✅ **Correto:**
```go
logger.Info("Request processed",
    "statusCode", statusCode,
    "duration", duration.Milliseconds(),
    "ip", ip,
)
```

❌ **Incorreto:**
```go
logger.Info("Request processed")
```

### 2. Use nomes de chaves consistentes
✅ **Correto:**
```go
logger.Error("Connection failed",
    "host", host,
    "port", port,
    "error", err,
)
```

❌ **Incorreto:**
```go
logger.Error("Connection failed",
    "h", host,
    "p", port,
    "err", err,
)
```

### 3. Para erros, sempre inclua o erro
✅ **Correto:**
```go
if err != nil {
    logger.Error("Failed to process request", "error", err)
    return err
}
```

❌ **Incorreto:**
```go
if err != nil {
    logger.Error(fmt.Sprintf("Failed to process request: %v", err))
    return err
}
```

### 4. Use debug para informações detalhadas
```go
logger.Debug("Cache hit",
    "key", cacheKey,
    "ttl", ttl.Seconds(),
    "size", size,
)
```

### 5. Use warn para situações recuperáveis
```go
logger.Warn("Slow query detected",
    "query", query,
    "duration", duration.Milliseconds(),
    "threshold", thresholdMs,
)
```

## Configuração de Níveis

### Produção (apenas Info, Warn, Error)
```go
// Em pkg/logger/logger.go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
```

### Desenvolvimento (Debug ativado)
```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
```

### Troubleshooting (Todos os logs)
```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
```

## Configuração por Variável de Ambiente (Futura Melhoria)

```go
// Adicionar ao pkg/logger/logger.go
func init() {
    logLevel := os.Getenv("LOG_LEVEL")
    level := slog.LevelInfo
    
    switch logLevel {
    case "debug":
        level = slog.LevelDebug
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    }
    
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    })
    defaultLogger = slog.New(handler)
    slog.SetDefault(defaultLogger)
}
```

Uso:
```bash
LOG_LEVEL=debug go run main.go
LOG_LEVEL=warn go run main.go
```

## Exemplo Prático: Middleware HTTP

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        startTime := time.Now()
        
        // Log request
        logger.Debug("Incoming request",
            "method", r.Method,
            "path", r.URL.Path,
            "ip", getClientIP(r),
        )
        
        // Call next handler
        next.ServeHTTP(w, r)
        
        // Log response
        duration := time.Since(startTime)
        logger.Info("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "duration_ms", duration.Milliseconds(),
        )
    })
}
```

## Exemplo Prático: Rate Limiter Customizado

```go
func (rl *RateLimiter) CheckLimit(ctx context.Context, id string) (bool, error) {
    logger.Debug("Checking rate limit", "id", id)
    
    allowed, err := rl.storage.CheckAndIncrement(ctx, id, rl.config.MaxRequests, 60)
    if err != nil {
        logger.Error("Rate limit check failed",
            "id", id,
            "error", err,
        )
        return false, err
    }
    
    if !allowed {
        logger.Warn("Rate limit exceeded",
            "id", id,
            "limit", rl.config.MaxRequests,
        )
        return false, nil
    }
    
    logger.Debug("Request allowed", "id", id)
    return true, nil
}
```

## Output em Diferentes Ambientes

### Desenvolvimento (stderr colorido - futura melhoria)
```
[DEBUG] Checking rate limit id=user123
[INFO] Starting server address=:8080
[WARN] Slow query detected query=SELECT* duration=1234
[ERROR] Connection failed host=db.local error=timeout
```

### Produção (JSON estruturado)
```json
{"time":"2024-12-11T10:30:50.456789012Z","level":"INFO","msg":"Request allowed","path":"/api/users","method":"GET","ip":"192.168.1.100","hasToken":false}
{"time":"2024-12-11T10:30:55.678901234Z","level":"WARN","msg":"Rate limit exceeded","path":"/api/users","ip":"192.168.1.100","blockDuration":60}
```

## Integração com Docker

### Dockerfile
```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go build -o rate-limiter

# Configurar log level via variável de ambiente
ENV LOG_LEVEL=info

EXPOSE 8080
CMD ["./rate-limiter"]
```

### docker-compose.yml
```yaml
services:
  rate-limiter:
    build: .
    environment:
      LOG_LEVEL: debug  # ou info, warn, error
      REDIS_ADDR: redis:6379
    ports:
      - "8080:8080"
    depends_on:
      - redis
```

## Monitoramento e Análise

### Usando jq para análise local
```bash
# Filtrar apenas erros
go run main.go 2>&1 | jq 'select(.level == "ERROR")'

# Contar eventos por tipo
go run main.go 2>&1 | jq -r '.msg' | sort | uniq -c | sort -rn

# Mostrar apenas warnings e errors
go run main.go 2>&1 | jq 'select(.level == "WARN" or .level == "ERROR")'
```

### Com ferramentas de stream
```bash
# Mostrar logs em tempo real com formato legível
go run main.go 2>&1 | jq -r '\(.time) [\(.level)] \(.msg) \(.ip // "N/A")'
```

## Boas Práticas Finais

1. ✅ **Inclua contexto específico** - IP, UserID, RequestID, Duration
2. ✅ **Use nomes de chaves descritivas** - `userID`, `duration_ms`, não `u`, `d`
3. ✅ **Estruture valores complexos** - Não use string para objetos
4. ✅ **Log em pontos críticos** - Inicio/fim de operações importantes
5. ✅ **Não log dados sensíveis** - Senhas, tokens, dados PII
6. ✅ **Use níveis apropriados** - DEBUG para diagnóstico, INFO para eventos
7. ✅ **Inclua erros completos** - `"error", err` captura stack trace em produção
8. ✅ **Evite logs desnecessários** - Não log em loops intensivos
