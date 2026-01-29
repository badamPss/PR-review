# PR Reviewer Assignment Service

Сервис, который назначает ревьюеров на PR из команды автора, позволяет выполнять переназначение ревьюверов и получать список PR’ов, назначенных конкретному пользователю, а также управлять командами и активностью пользователей. 

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

Обновленная схема и примеры — в `docs/openapi.yml`.

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

Покрытие интеграционными тестами - 71%:
```bash
go test -v ./internal/integration \
  -cover -coverpkg=./internal/... \
  -coverprofile=it.cover
```

## Нагрузочное тестирование

Нагрузочное тестирование выполнено с помощью Apache Bench (ab).

### Запуск тестов

```bash
./scripts/load_test.sh
```

Параметры можно настроить через переменные окружения:
```bash
REQUESTS=500 CONCURRENCY=10 ./scripts/load_test.sh
```

### Результаты

**Целевые SLI из ТЗ:**
- RPS: 5
- Время ответа: ≤ 300 мс
- Успешность: ≥ 99.9%

**Фактические результаты:**

| Эндпоинт | RPS | Latency (mean) | Latency (p95) | Latency (p99) | Статус |
|----------|-----|----------------|---------------|---------------|--------|
| `GET /stats` | ~3,889 | 1.3 мс | 2 мс | 6 мс | ✅ |
| `GET /users/getReview` | ~6,138 | 0.8 мс | 1 мс | 1 мс | ✅ |
| `GET /team/get` | ~6,793 | 0.7 мс | 1 мс | 3 мс | ✅ |
| `POST /team/add` | ~6,915 | 0.7 мс | 1 мс | 2 мс | ✅ |
| `POST /team/deactivateMembers` | ~5,629 | 0.9 мс | 1 мс | 2 мс | ✅ |
| `POST /users/setIsActive` | ~4,402 | 1.1 мс | 2 мс | 9 мс | ✅ |
| `POST /pullRequest/create` | ~63 | 16 мс | 24 мс | 32 мс | ✅ |
| `POST /pullRequest/merge` | ~7,406 | 0.7 мс | 1 мс | 1 мс | ✅ |
| `POST /pullRequest/reassign` | ~6,672 | 0.7 мс | 1 мс | 1 мс | ✅ |

**Заключение:** Все 9 эндпоинтов **значительно превышают** требования:
- **RPS**: 63-7,400 RPS (требование: 5 RPS) - в 12-1481 раз выше
- **Latency**: 0.7-16 мс (требование: ≤ 300 мс) - в 18-428 раз лучше
- **Успешность**: Все запросы успешно обрабатываются

Подробный отчет: `load_test_results/SUMMARY.md`
