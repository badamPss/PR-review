# PR Reviewer Assignment Service (Test Task, Fall 2025)

Короткий README по запуску и проверке сервиса, реализующего автоматическое назначение ревьюверов и управление командами/пользователями.

## Стек и решения
- Язык: Go
- БД: PostgreSQL
- Пользователь принадлежит ровно одной команде: `users.team_id`
- Миграции: SQL (`db/migration`), применяются при старте контейнером `migrate`
- HTTP: Echo; соответствие спецификации `openapi.yml`

## Запуск
Требуется Docker и Docker Compose.

```bash
docker-compose up -d
```

Альтернатива через Makefile (короткие команды):
```bash
make up        # поднять
make logs      # логи
make down      # остановить
```

Сервис доступен на: `http://localhost:8080`.

## Миграции
При `docker-compose up` автоматически применяются `*.up.sql` из `db/migration`. Для чистого старта:
```bash
docker-compose down -v && docker-compose up -d
```

## Основные эндпоинты (без префиксов)
- `POST /team/add` — создать команду с участниками (создаёт/обновляет пользователей)
- `GET /team/get?team_name=...` — получить команду
- `POST /team/deactivateMembers` — деактивировать всех пользователей команды и удалить их из списка ревьюверов открытых PR
- `POST /users/setIsActive` — включить/выключить активность пользователя
- `POST /pullRequest/create` — создать PR и автоматически назначить до 2 активных ревьюверов из команды автора (кроме автора)
- `POST /pullRequest/merge` — отметить PR как MERGED (идемпотентно)
- `POST /pullRequest/reassign` — переназначить ревьювера на случайного активного из команды заменяемого
- `GET /users/getReview?user_id=...` — PR'ы, где пользователь назначен ревьювером
- `GET /stats` — статистика: назначения по пользователям и число ревьюверов по PR

## Пример запроса статистики
```bash
curl -s http://localhost:8080/stats | jq
```
Ответ:
```json
{
  "by_user": [
    { "user_id": "u2", "assignments": 3 }
  ],
  "per_pr": [
    { "pull_request_id": "pr-1001", "reviewers_count": 2 }
  ]
}
```

## Пример деактивации команды
```bash
curl -s -X POST http://localhost:8080/team/deactivateMembers \
  -H "Content-Type: application/json" \
  -d '{"team_name": "backend"}' | jq
```
Ответ:
```json
{
  "team_name": "backend",
  "reassigned_prs_count": 1
}
```

**Что происходит:**
1. Все пользователи команды деактивируются (`is_active = false`)
2. Все открытые PR, где есть ревьюверы из этой команды, обновляются — деактивированные ревьюверы удаляются из списка
3. Возвращается количество обновлённых PR

Схема и примеры — в `openapi.yml`.

## Линтер
В проекте используется `golangci-lint` v2, конфиг — `.golangci.yml`.

Запуск локально (если установлен golangci-lint):
```bash
make lint
```

## Интеграционные тесты
Интеграционные тесты находятся в `internal/integration`: сервис поднимается в памяти (httptest) с реальной PostgreSQL через testcontainers.

Запуск:
```bash
make test-integration
```

Покрытие интеграционными тестами - 70%:
```bash
go test -v ./internal/integration \
  -cover -coverpkg=./internal/... \
  -coverprofile=it.cover
```