# Padrão de Logging Estruturado

## Overview

O projeto implementa um padrão centralizado de logging estruturado utilizando o pacote `log/slog` do Go (disponível desde Go 1.21). Isso facilita o monitoramento, debugging e análise de logs em produção.

## Estrutura

### Pacote de Logger

O pacote centralizado está localizado em [pkg/logger/logger.go](../../pkg/logger/logger.go) e fornece as seguintes funções:

- **`Info(msg string, args ...any)`** - Logs informativos com pares de chave-valor
- **`Warn(msg string, args ...any)`** - Logs de aviso
- **`Error(msg string, args ...any)`** - Logs de erro
- **`Debug(msg string, args ...any)`** - Logs de debug
- **`Fatal(msg string, args ...any)`** - Logs de erro crítico e saída do programa

### Formato de Saída

Os logs são emitidos em formato **JSON**, o que facilita:
- Parsing automático por ferramentas de observabilidade (DataDog, Grafana, etc.)
- Filtragem e busca eficiente
- Integração com sistemas de monitoramento

## Exemplo de Uso

### Antes (sem estrutura)
```go
fmt.Printf("Starting server with config:\n")
fmt.Printf("  IP Limit: %d req/s\n", cfg.MaxRequestsIP)
log.Fatalf("Failed to initialize Redis: %v", err)
```

### Depois (estruturado)
```go
logger.Info("Starting rate limiter server",
    "maxRequestsIP", cfg.MaxRequestsIP,
    "enableIPLimit", cfg.EnableIPLimit,
    "maxRequestsToken", cfg.MaxRequestsToken,
    "enableTokenLimit", cfg.EnableTokenLimit,
    "redisAddr", cfg.RedisAddr,
)

if err != nil {
    logger.Fatal("Failed to initialize Redis", "error", err)
}
```

## Arquivos Atualizados

1. **[main.go](../../main.go)** - Logs de inicialização do servidor
2. **[config/loader.go](../../config/loader.go)** - Logs de carregamento de configuração
3. **[middleware/middleware.go](../../middleware/middleware.go)** - Logs de requisições bloqueadas e erros
4. **[limiter/limiter.go](../../limiter/limiter.go)** - Logs de validação de rate limit
5. **[storage/redis.go](../../storage/redis.go)** - Logs de conexão e operações Redis

## Boas Práticas Implementadas

### 1. Mensagens Clara e Concisa
```go
logger.Info("Server listening", "address", addr)
```

### 2. Contexto com Pares Chave-Valor
```go
logger.Warn("Rate limit exceeded",
    "path", r.RequestURI,
    "ip", ip,
    "blockDuration", blockDuration,
)
```

### 3. Sensibilidade de Dados
```go
logger.Debug("Configuration loaded", "REDIS_PASS", "***")
```

### 4. Níveis Apropriados
- **Info** - Eventos importantes do sistema (inicialização, eventos significativos)
- **Warn** - Situações inesperadas mas recuperáveis (rate limits, timeouts)
- **Error** - Erros que afetam funcionalidades
- **Debug** - Informações detalhadas para debugging
- **Fatal** - Erros críticos que impedem execução

## Exemplo de Saída JSON

```json
{
  "time": "2024-12-11T10:30:45.123456789Z",
  "level": "INFO",
  "msg": "Starting rate limiter server",
  "maxRequestsIP": 10,
  "enableIPLimit": true,
  "maxRequestsToken": 100,
  "enableTokenLimit": true,
  "redisAddr": "localhost:6379"
}
```

## Configuração Futura

Para ambientes de produção, você pode configurar diferentes níveis de log:

```go
// Em pkg/logger/logger.go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,  // Alterar para slog.LevelDebug, slog.LevelWarn, etc.
})
```

Ou via variável de ambiente (implementação futura):
```bash
export LOG_LEVEL=debug
```

## Ferramentas de Observabilidade

Os logs estruturados em JSON funcionam perfeitamente com:
- **DataDog** - Parsing automático de campos JSON
- **Grafana Loki** - Filtragem por labels estruturados
- **Elastic Stack** - Indexação automática de campos
- **CloudWatch** - Insights sobre padrões de logs
- **Splunk** - Busca avançada em logs estruturados
