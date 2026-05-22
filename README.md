# API организационной структуры

REST API для управления иерархией подразделений и сотрудников компании.  
Реализовано на **Go (net/http)** с использованием **GORM**, **PostgreSQL**, миграций **goose**, упаковано в **Docker** (docker-compose).  
Включает минимальный веб-интерфейс для демонстрации работы API.

---

## Быстрый старт

### 1. Склонируйте репозиторий

```bash
git clone https://github.com/tim44ik/hitalent-tt
cd orgstructure
```

### 2. Настройте окружение

Создайте файл `.env` (можно скопировать из примера):

```bash
cp .env.example .env
```

Отредактируйте при необходимости. Пример `.env`:

```
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hitalent-test
SERVER_PORT=8080
```

### 3. Запустите через Docker Compose

```bash
docker-compose up --build
```

После запуска:
- **API** доступно: `http://localhost:8080`
- **Веб-интерфейс** (UI) доступен по тому же адресу
- **Health checks**: `/live`, `/ready`, `/startup`

### 4. Остановка

```bash
docker-compose down
```

---

## API Endpoints

Базовый URL: `http://localhost:8080`

### 1. Создать подразделение

**POST** `/departments/`

Тело (JSON):
```json
{
  "name": "Название отдела",
  "parent_id": null   // или ID родителя
}
```

Ответ: `201 Created` – объект подразделения.

### 2. Создать сотрудника в подразделении

**POST** `/departments/{id}/employees`

Тело:
```json
{
  "full_name": "Иванов Иван Иванович",
  "position": "Разработчик",
  "hired_at": "2025-01-01"   // опционально
}
```

Ответ: `201 Created` – объект сотрудника.

### 3. Получить подразделение с деревом

**GET** `/departments/{id}`

Параметры строки запроса:
- `depth` – от 1 до 5 (по умолчанию 1) – глубина вложенных подразделений
- `include_employees` – `true`/`false` (по умолчанию `true`)

Пример:  
`GET /departments/2?depth=2&include_employees=true`

Ответ: `200 OK` – объект подразделения с полями `children[]` и `employees[]`.

### 4. Переместить / переименовать подразделение

**PATCH** `/departments/{id}`

Тело (оба поля опциональны):
```json
{
  "name": "Новое имя",
  "parent_id": 5   // или null – сделать корневым
}
```

Ответ: `200 OK` – обновлённое подразделение.

Ограничения:
- Нельзя сделать родителем самого себя или потомка (409 Conflict)
- Имя должно быть уникальным среди подразделений того же родителя (409 Conflict)

### 5. Удалить подразделение

**DELETE** `/departments/{id}`

Параметры строки запроса:
- `mode` – `cascade` (удалить всё поддерево) или `reassign` (перевести сотрудников и детей)
- `reassign_to_department_id` – обязателен при `mode=reassign`

Примеры:
- `DELETE /departments/3?mode=cascade`
- `DELETE /departments/3?mode=reassign&reassign_to_department_id=1`

Ответ: `204 No Content` (без тела).

---

## Тестирование

### Unit-тесты (сервисы, моки)

```bash
go test -v ./tests/unit/...
```

### Интеграционный тест (API + SQLite in‑memory)

```bash
go test -v ./tests/integration/...
```

> Для полноценного тестирования с PostgreSQL можно использовать **testcontainers** (код приведён в комментариях к тестам).

---

## 📁 Архитектура проекта

```
.
├── cmd/app/                 # точка входа main
├── internal/
│   ├── app/                 # сборка приложения, graceful shutdown, ready‑флаг
│   ├── config/              # загрузка переменных окружения
│   ├── handlers/            # HTTP‑обработчики, DTO, ответы
│   ├── middleware/          # логирование, recovery
│   ├── models/              # GORM‑модели (только поля БД + gorm:"-")
│   ├── repositories/        # интерфейсы и реализация доступа к данным
│   ├── server/              # роутер (chi) с middleware и пробросами
│   ├── services/            # бизнес‑логика (валидация, уникальность, циклы)
│   └── ui/                  # встроенный HTML‑интерфейс (embed)
├── migrations/              # SQL‑миграции goose
├── tests/                   # unit и интеграционные тесты
├── Dockerfile
├── docker-compose.yml
├── entrypoint.sh            # запуск goose up, затем приложения
├── .env.example
└── README.md
```

Принципы:
- Чистая архитектура (слои изолированы)
- Использование интерфейсов для тестируемости
- Все ошибки сервисов – отдельные переменные (`errors.New`), сравнение через `errors.Is`
- Конфигурация только через переменные окружения, секреты в `.env` (не в репозитории)

---

## Docker

- **Dockerfile**: многостадийная сборка (builder на `golang:1.23-alpine`, runtime на `alpine`). Бинарный файл `app` копируется в финальный образ.
- **entrypoint.sh**: перед запуском `app` выполняет `goose up` для миграций БД.
- **docker-compose.yml**:
  - Сервис `postgres` – официальный образ, healthcheck через `pg_isready`.
  - Сервис `app` – ждёт здорового postgres, монтирует переменные из `.env`, имеет healthcheck по `/ready`.
- **Health-пробы**:
  - `/live` – всегда 200 (процесс жив)
  - `/ready` – 200 после завершения инициализации (БД подключена, миграции выполнены)
  - `/startup` – аналогичен `/ready`

---

## Локальная разработка (без Docker)

Требуется PostgreSQL (локально установленный) и Go 1.23+.

```bash
# Установка goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Накатить миграции
goose -dir migrations postgres "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Запуск приложения
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=orgstructure
go run cmd/app/main.go
```
