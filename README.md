# PR Reviewer Assignment Service (Test Task, Fall 2025)

Короткий README по запуску и проверке сервиса, реализующего автоматическое назначение ревьюверов и управление командами/пользователями.

## Стек и решения
- Язык: Go
- БД: PostgreSQL
- Идентификаторы (user_id, pull_request_id) — строки (соответствие OpenAPI)
- Пользователь принадлежит ровно одной команде: `users.team_id` (таблица `team_members` не используется)
- Миграции: SQL (`db/migration`), применяются при старте контейнером `migrate`
- HTTP: Echo; соответствие спецификации `openapi.yml`

## Запуск
Требуется Docker и Docker Compose.

```bash
docker-compose up -d
# логи
docker-compose logs -f
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

## Быстрый сценарий проверки (curl)
1) Создать команду:
```bash
curl -s -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{"team_name":"backend","members":[
    {"user_id":"u1","username":"Alice","is_active":true},
    {"user_id":"u2","username":"Bob","is_active":true},
    {"user_id":"u3","username":"Charlie","is_active":true}
  ]}'
```
2) Создать PR (назначение ревьюверов):
```bash
curl -s -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr-1","pull_request_name":"Init","author_id":"u1"}'
```
3) Переназначить одного из назначенных ревьюверов:
```bash
curl -s -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr-1","old_user_id":"u2"}'
```
4) Идемпотентный merge:
```bash
curl -s -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr-1"}'
```

## Ошибки (ErrorResponse.error.code)
- `TEAM_EXISTS`, `PR_EXISTS`, `PR_MERGED`, `NOT_ASSIGNED`, `NO_CANDIDATE`, `NOT_FOUND`

## Нефункциональные требования
- Объёмы: ≤ 20 команд и ≤ 200 пользователей
- RPS ~ 5, SLI 300 мс / 99.9% — соблюдается:
  - Простые индексы (`users(team_id)`, `users(is_active)`, `pull_requests(pull_request_id)`, `status` и т.д.)
  - Запросы точечные, без тяжёлых операций

## Принятые допущения
- У пользователя ровно одна команда (поле `team_name`/`team_id`), т.к. так отражено в спецификации.
- Если кандидатов < 2, назначается доступное кол-во (0/1).
- После `MERGED` список ревьюверов менять нельзя (возвращается `PR_MERGED`).

## Полезное
- Коллекция Postman/окружение можно легко собрать из `openapi.yml`, либо пользоваться командами curl из раздела выше.

# Сервис назначения ревьюеров для Pull Request'ов

Внутри команды требуется единый микросервис, который автоматически назначает ревьюеров на Pull Request'ы (PR), а также позволяет управлять командами и участниками. Взаимодействие происходит исключительно через HTTP API.

## Общие вводные

**Пользователь (User)** — участник команды с уникальным идентификатором, именем и флагом активности `isActive`.

**Команда (Team)** — группа пользователей с уникальным именем.

**Pull Request (PR)** — сущность с идентификатором, названием, автором, статусом `OPEN|MERGED`и списком назначенных ревьюверов (до 2).

1. При создании PR автоматически назначаются **до двух** активных ревьюверов из **команды автора**, исключая самого автора.

2. Переназначение заменяет одного ревьювера на случайного **активного** участника **из команды заменяемого** ревьювера.

3. После `MERGED` менять список ревьюверов **нельзя**.

4. Если доступных кандидатов меньше двух, назначается доступное количество (0/1).

## Запуск проекта

### Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)

### Запуск через Docker Compose

Самый простой способ запустить проект:

```bash
make docker-start
```

Или вручную:

```bash
docker-compose -f docker-compose-local.yaml up --build
```

Сервис будет доступен на порту `8080`.

### Локальная разработка

1. Установите зависимости:
```bash
go mod download
```

2. Запустите PostgreSQL (через Docker):
```bash
docker-compose -f docker-compose-local.yaml up -d postgres
```

3. Примените миграции:
```bash
make migrate-up
```

4. Запустите приложение:
```bash
make run
```

## API Endpoints

Все эндпоинты доступны по префиксу `/api/v1`:

- `POST /api/v1/team/add` - Создать команду с участниками
- `GET /api/v1/team/get?team_name=...` - Получить команду с участниками
- `POST /api/v1/users/setIsActive` - Установить флаг активности пользователя
- `GET /api/v1/users/getReview?user_id=...` - Получить PR'ы, где пользователь назначен ревьювером
- `POST /api/v1/pullRequest/create` - Создать PR и автоматически назначить ревьюверов
- `POST /api/v1/pullRequest/merge` - Пометить PR как MERGED
- `POST /api/v1/pullRequest/reassign` - Переназначить ревьювера

Полная спецификация API доступна в файле `openapi.yml`.

## Makefile команды

- `make build` - Собрать приложение
- `make run` - Запустить приложение локально
- `make test` - Запустить тесты
- `make clean` - Очистить артефакты сборки
- `make docker-build` - Собрать Docker образ
- `make docker-up` - Запустить все сервисы
- `make docker-down` - Остановить все сервисы
- `make docker-start` - Полный запуск (сборка + запуск)
- `make migrate-up` - Применить миграции
- `make migrate-down` - Откатить миграции
- `make docker-logs` - Просмотр логов

## Структура проекта

```
.
├── cmd/pr-review/     # Точка входа приложения
├── internal/
│   ├── app/          # Инициализация приложения
│   ├── config/       # Конфигурация
│   ├── handlers/     # HTTP handlers
│   ├── models/       # Модели данных
│   ├── repository/   # Слой доступа к данным
│   └── service/      # Бизнес-логика
├── db/migration/     # Миграции БД
├── config/           # Файлы конфигурации
└── docker/           # Docker конфигурации
```

## Принятые решения

1. **ID пользователей**: Используются строковые ID из API, которые конвертируются в int64 для хранения в БД
2. **Автоназначение ревьюверов**: Реализовано случайное назначение до 2 активных ревьюверов из команды автора
3. **Переназначение**: Новый ревьювер выбирается из команды заменяемого ревьювера
4. **Идемпотентность merge**: Повторный вызов merge возвращает текущее состояние PR без ошибки
5. **Обработка ошибок**: Все ошибки возвращаются в формате согласно OpenAPI спецификации

## Технологии

- Go 1.25
- PostgreSQL
- Echo (HTTP framework)
- SQLx (database driver)
- Squirrel (SQL query builder)
- Docker & Docker Compose
