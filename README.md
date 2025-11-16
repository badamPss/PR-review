# PR Reviewer Assignment Service (Test Task, Fall 2025)

Короткий README по запуску и проверке сервиса, реализующего автоматическое назначение ревьюверов и управление командами/пользователями.

## Стек и решения
- Язык: Go
- БД: PostgreSQL
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
- `POST /users/setIsActive` — включить/выключить активность пользователя
- `POST /pullRequest/create` — создать PR и автоматически назначить до 2 активных ревьюверов из команды автора (кроме автора)
- `POST /pullRequest/merge` — отметить PR как MERGED (идемпотентно)
- `POST /pullRequest/reassign` — переназначить ревьювера на случайного активного из команды заменяемого
- `GET /users/getReview?user_id=...` — PR’ы, где пользователь назначен ревьювером

Схема и примеры — в `openapi.yml`.
