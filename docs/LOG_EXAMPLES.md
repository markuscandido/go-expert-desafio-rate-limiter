# Exemplos de Logs Estruturados

Este arquivo demonstra os logs que serão gerados com o novo padrão de logging estruturado.

## Inicialização da Aplicação

### Carregamento de Configuração
```json
{
  "time": "2024-12-11T10:30:44.987654321Z",
  "level": "DEBUG",
  "msg": "Loading configuration from environment"
}

{
  "time": "2024-12-11T10:30:44.987654321Z",
  "level": "DEBUG",
  "msg": "Configuration loaded",
  "RATE_LIMITER_ENABLE_IP": true
}

{
  "time": "2024-12-11T10:30:44.987654321Z",
  "level": "INFO",
  "msg": "Configuration loaded successfully",
  "ipLimitEnabled": true,
  "tokenLimitEnabled": true
}
```

### Conexão com Redis
```json
{
  "time": "2024-12-11T10:30:45.123456789Z",
  "level": "INFO",
  "msg": "Connected to Redis",
  "addr": "localhost:6379",
  "db": 0
}
```

### Servidor Iniciado
```json
{
  "time": "2024-12-11T10:30:45.234567890Z",
  "level": "INFO",
  "msg": "Starting rate limiter server",
  "maxRequestsIP": 10,
  "enableIPLimit": true,
  "maxRequestsToken": 100,
  "enableTokenLimit": true,
  "redisAddr": "localhost:6379"
}

{
  "time": "2024-12-11T10:30:45.345678901Z",
  "level": "INFO",
  "msg": "Server listening",
  "address": ":8080"
}
```

## Operações Normais

### Requisição Permitida
```json
{
  "time": "2024-12-11T10:30:50.456789012Z",
  "level": "DEBUG",
  "msg": "Request allowed",
  "path": "/api/users",
  "method": "GET",
  "ip": "192.168.1.100",
  "hasToken": false
}
```

### Requisição com Token Permitida
```json
{
  "time": "2024-12-11T10:30:51.567890123Z",
  "level": "DEBUG",
  "msg": "Request allowed",
  "path": "/api/data",
  "method": "POST",
  "ip": "203.0.113.42",
  "hasToken": true
}
```

## Situações de Rate Limit

### Rate Limit Excedido por IP
```json
{
  "time": "2024-12-11T10:30:55.678901234Z",
  "level": "WARN",
  "msg": "Rate limit exceeded",
  "path": "/api/users",
  "ip": "192.168.1.100",
  "hasToken": false,
  "blockDuration": 60
}
```

### Rate Limit Excedido por Token
```json
{
  "time": "2024-12-11T10:30:56.789012345Z",
  "level": "WARN",
  "msg": "Rate limit exceeded",
  "path": "/api/premium",
  "ip": "203.0.113.99",
  "hasToken": true,
  "blockDuration": 120
}
```

### IP Bloqueado
```json
{
  "time": "2024-12-11T10:30:57.890123456Z",
  "level": "WARN",
  "msg": "IP blocked",
  "ip": "192.168.1.100",
  "blockDuration": 60
}
```

## Operações Redis

### Limite Atingido (Threshold)
```json
{
  "time": "2024-12-11T10:30:58.901234567Z",
  "level": "DEBUG",
  "msg": "Rate limit threshold reached",
  "key": "ip:192.168.1.100",
  "count": 11,
  "maxRequests": 10
}
```

### Chave Bloqueada
```json
{
  "time": "2024-12-11T10:30:59.012345678Z",
  "level": "DEBUG",
  "msg": "Key blocked",
  "key": "ip:192.168.1.100",
  "durationSeconds": 60
}
```

### Reset de Chave
```json
{
  "time": "2024-12-11T10:31:00.123456789Z",
  "level": "DEBUG",
  "msg": "Key reset",
  "key": "ip:192.168.1.100"
}
```

## Situações de Erro

### Erro ao Conectar ao Redis
```json
{
  "time": "2024-12-11T10:30:45.234567890Z",
  "level": "ERROR",
  "msg": "Failed to connect to Redis",
  "addr": "localhost:6379",
  "error": "connection refused"
}
```

### Erro ao Obter Dados do Redis
```json
{
  "time": "2024-12-11T10:30:50.345678901Z",
  "level": "ERROR",
  "msg": "Failed to get data from Redis",
  "key": "ip:192.168.1.100",
  "error": "WRONGTYPE Operation against a key holding the wrong kind of value"
}
```

### Erro ao Verificar Rate Limit
```json
{
  "time": "2024-12-11T10:30:51.456789012Z",
  "level": "ERROR",
  "msg": "Rate limiter error",
  "path": "/api/users",
  "ip": "192.168.1.100",
  "hasToken": false,
  "error": "context deadline exceeded"
}
```

### Erro de Configuração
```json
{
  "time": "2024-12-11T10:30:44.567890123Z",
  "level": "WARN",
  "msg": "Invalid value for RATE_LIMITER_MAX_REQUESTS_IP",
  "value": "abc",
  "error": "strconv.Atoi: parsing \"abc\": invalid syntax"
}
```

## Encerramento

### Fechamento de Conexão Redis
```json
{
  "time": "2024-12-11T10:35:00.678901234Z",
  "level": "INFO",
  "msg": "Closing Redis connection"
}
```

### Erro Crítico no Servidor
```json
{
  "time": "2024-12-11T10:35:01.789012345Z",
  "level": "ERROR",
  "msg": "Server error",
  "error": "listen tcp :8080: bind: address already in use"
}
```

## Filtragem e Análise com Ferramentas

### Encontrar todos os rate limits excedidos:
```bash
grep "Rate limit exceeded" logs.json
jq 'select(.msg == "Rate limit exceeded")' logs.jsonl
```

### Encontrar erros:
```bash
jq 'select(.level == "ERROR")' logs.jsonl | jq '.msg, .error'
```

### Estatísticas por IP:
```bash
jq -r '.ip' logs.jsonl | sort | uniq -c | sort -rn
```

### Requisições bloqueadas em um período:
```bash
jq 'select(.msg == "Rate limit exceeded" and .time > "2024-12-11T10:30:00Z")' logs.jsonl
```

## Integração com Observabilidade

### Com Grafana Loki:
```promql
{job="rate-limiter", level="WARN"} |= "Rate limit exceeded"
```

### Com DataDog:
```
service:rate-limiter level:WARN msg:"Rate limit exceeded"
```

### Com Elastic Stack:
```json
{
  "query": {
    "bool": {
      "must": [
        { "match": { "msg": "Rate limit exceeded" } },
        { "range": { "time": { "gte": "2024-12-11T10:00:00Z" } } }
      ]
    }
  }
}
```
