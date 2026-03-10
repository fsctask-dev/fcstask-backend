# Пакет middleware
## internal/middleware/
```text
middleware/
├── metrics.go   # Сбор HTTP метрик (запросы, статусы, время)
└── inflight.go  # Активные запросы в полёте
```

## MetricsMiddleware
Автоматически собирает метрики для всех HTTP запросов.

```go
metricsMW := middleware.NewMetricsMiddleware(client)

// Оборачиваем роутер
handler := metricsMW.Handler(mux)
http.ListenAndServe(":8080", handler)
```

Собирает:

| Метод                         | Имя метрики                    | Теги                               |
|-------------------------------|--------------------------------|------------------------------------|
| http_requests_total           | Counter                        | path, method, status              |
| http_request_duration_seconds | Histogram                      | path, method                      |
Пример тегов:

```json
{
"path": "/api/v1/users",
"method": "POST",
"status": "201"
}
```

## InflightMetricsMiddleware
Считает количество одновременно обрабатываемых запросов.

```go
inflightMW := middleware.NewInflightMetricsMiddleware(client)
handler := inflightMW.Handler(mux)
```
Собирает:

| Метрика                   | Тип    | Описание                     |
|---------------------------|--------|------------------------------|
| http_requests_in_flight   | Gauge  | Текущие активные запросы     |

Особенности:

* Увеличивается на 1 в начале запроса

* Уменьшается на 1 в defer

* Работает даже при панике

## Композиция middleware
```go
// Порядок важен: inflight → metrics → handlers
handler := inflightMW.Handler(metricsMW.Handler(mux))
```
