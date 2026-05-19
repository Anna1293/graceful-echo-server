# graceful-echo-server

Минимальный Go-сервис: backend echo-сервер и reverse proxy на `net/http` с TLS и HTTP/2.

## Возможности

- Эндпоинт `GET /echo` с необязательными параметрами сообщения и задержки
- Обработка отмены запроса через `r.Context().Done()`
- Корректное завершение по `SIGINT` и `SIGTERM`
- Reverse proxy на Go (без Nginx) через стандартный `net/http`
- TLS-терминация и HTTP/2 на входящем HTTPS-трафике

## Требования

- Go 1.25+

## Запуск

```bash
go run .
```

Сервис поднимает два сервера:
- backend (`http`) на `:8080`;
- reverse proxy (`https`) на `:8443` с поддержкой HTTP/2.

## Использование

### Простой запрос

```bash
curl "http://localhost:8080/echo?msg=hello&delay=3"
```

Ответ:

```text
echo: hello
```

### Значения по умолчанию

Если параметр `msg` не передан, сервер отвечает `echo`.

Если параметр `delay` не передан, сервер использует `5` секунд.

### Валидация

`delay` должен быть неотрицательным целым числом. Иначе сервер возвращает `400 Bad Request`.

## Graceful shutdown

1. Запусти сервер: `go run .`
2. Отправь длинный запрос (например, с `delay=10`)
3. Останови сервер через `Ctrl+C`

Оба сервера (backend и proxy) перестают принимать новые запросы и ждут до 10 секунд завершения активных запросов.

## Структура проекта

| Пакет | Назначение |
|-------|------------|
| `main.go` | Запуск backend и TLS reverse proxy, graceful shutdown |
| `internal/echo` | Echo backend (`GET /echo`) |
| `internal/proxy` | Reverse proxy (замена Nginx), заголовки `X-Forwarded-*` |
| `internal/certs` | Self-signed TLS при первом запуске |
| `internal/config` | Адреса и пути к сертификатам (env) |

## TLS и HTTP/2 для reverse proxy

При первом `go run .` сертификаты создаются автоматически в `certs/` (если файлов ещё нет). Либо вручную:

```powershell
.\scripts\generate-certs.ps1
```

Файлы: `certs/server.crt`, `certs/server.key` (в `.gitignore`).

### Запустить приложение

```bash
go run .
```

### Проверить HTTPS + HTTP/2

```bash
curl -k --http2 "https://localhost:8443/echo?msg=hello&delay=1"
```

Опция `-k` нужна, потому что сертификат самоподписанный.
