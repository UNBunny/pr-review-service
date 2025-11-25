# PR Review Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы.

## Технологии

Go 1.22+ • PostgreSQL 15 • Docker

## Быстрый старт

### Запуск сервиса

```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`

### Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Получить команду
curl http://localhost:8080/team/get?team_name=backend

# Статистика
curl http://localhost:8080/stats
```

## API Endpoints

### Команды

**POST /team/add** - Создать команду
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true}
    ]
  }'
```

**GET /team/get?team_name=<name>** - Получить команду

**POST /team/deactivate-all?team_name=<name>** - Деактивировать всех участников

### Пользователи

**POST /users/setIsActive** - Изменить статус пользователя
```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u1",
    "is_active": false
  }'
```

**GET /users/getReview?user_id=<id>** - Получить PR'ы пользователя как ревьюера

### Pull Requests

**POST /pullRequest/create** - Создать PR с автоназначением ревьюеров
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add new feature",
    "author_id": "u1"
  }'
```

**POST /pullRequest/merge** - Смержить PR (идемпотентно)
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-1001"}'
```

**POST /pullRequest/reassign** - Переназначить ревьюера
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }'
```

### Дополнительно

**GET /stats** - Статистика назначений  
**GET /health** - Проверка здоровья сервиса

## Как работает

- При создании PR автоматически назначаются до 2-х активных ревьюеров из команды автора (автор исключается)
- Переназначение заменяет ревьюера на случайного активного из его команды
- После MERGED изменения запрещены
- Мерж идемпотентный - повторный вызов возвращает 200 OK

## Бонусные задания

✅ **Статистика** - `GET /stats` показывает топ по количеству ревью  
✅ **Массовая деактивация** - `/team/deactivate-all` деактивирует команду и переназначает открытые PR  
✅ **Интеграционные тесты** - `make test` запускает E2E тесты  
✅ **Нагрузочное тестирование** - `make loadtest` для проверки под нагрузкой  
✅ **Линтер** - настроен `.golangci.yml`, запуск через `make lint`

## Принятые решения

**Случайный выбор ревьюеров** - через Fisher-Yates shuffle для равномерного распределения

**Идемпотентность мержа** - повторный вызов не приводит к ошибке (по условию задачи)

**Переназначение при деактивации** - ищем замену из команды автора PR, т.к. они знают контекст

**Clean Architecture** - разделил на слои (domain, service, repository, transport) для удобства тестирования

**Индексы** - добавил на `team_name`, `is_active`, `author_id`, `status`, `user_id` для быстрых запросов

## Makefile команды

```bash
make run       # Запустить через docker-compose
make build     # Собрать бинарник
make test      # Запустить тесты
make lint      # Проверить линтером
make clean     # Очистить артефакты и volumes
make loadtest  # Нагрузочное тестирование
make dev       # Запустить локально (только для разработки)
```
