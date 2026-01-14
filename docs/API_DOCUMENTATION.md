# Marketplace Platform

## Cтарт

### Docker Compose

```bash
# Клонировать репозиторий
git clone https://github.com/avito-edu/Thorlik_final_project
cd marketplace

# Запустить все сервисы
docker-compose up -d

# Проверить статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f app
```

Приложение будет доступно по адресу: http://localhost:8080


## Конфигурация

Приложение настраивается через переменные окружения или файл `.env`.

### Переменные окружения

| Переменная         | Описание                 | Значение по умолчанию |
|--------------------|--------------------------|-----------------------|
| `SERVER_HOST`      | Хост сервера             | `localhost`           |
| `SERVER_PORT`      | Порт сервера             | `8080`                |
| `DB_HOST`          | Хост PostgreSQL          | `localhost`           |
| `DB_PORT`          | Порт PostgreSQL          | `5432`                |
| `DB_USER`          | Пользователь БД          | `user123`             |
| `DB_PASSWORD`      | Пароль БД                | `marketplace123`      |
| `DB_NAME`          | Имя базы данных          | `marketplace_db`      |
| `DB_SSLMODE`       | SSL режим                | `disable`             |
| `JWT_SECRET`       | Секретный ключ JWT       | `secret-key`          |
| `JWT_EXPIRY_HOURS` | Время жизни токена (часы)| `24`                  |
| `LOG_LEVEL`        | Уровень логирования      | `info`                |

### Пример файла .env

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=user123
DB_PASSWORD=marketplace123
DB_NAME=marketplace_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=secret-key
JWT_EXPIRY_HOURS=24

# Logging
LOG_LEVEL=info
```

## Миграции базы данных

### Структура миграций

Миграции находятся в папке `migrations/`:

```
migrations/
├── 001_create_tables.up.sql    # Создание таблиц
└── 001_create_tables.down.sql  # Откат миграций
```

### Применение миграций

```bash
# Применить миграции (создание таблиц)
psql -U marketplace -d marketplace_db -f migrations/001_create_tables.up.sql

# Откатить миграции (удаление таблиц)
psql -U marketplace -d marketplace_db -f migrations/001_create_tables.down.sql
```

### Таблицы

| Таблица       | Описание                                          |
|---------------|---------------------------------------------------|
| `users`       | Пользователи системы (юзеры, модераторы, админы)  |
| `products`    | Товары на продажу                                 |
| `orders`      | Заказы пользователей                              |
| `order_items` | Позиции заказа (связь заказ-товар)                |
| `cart_items`  | Корзина покупок                                   |

### Тестовый пользователь

После применения миграций создаётся администратор:

- **Email:** `admin@marketplace.com`
- **Password:** `admin123`

---

## API документация

### Swagger UI

После запуска приложения Swagger доступен по адресу:

**http://localhost:8080/swagger/index.html**

### Перегенерация Swagger

При изменении аннотаций в коде:

```bash
# Установка swag (если не установлен)
go install github.com/swaggo/swag/cmd/swag@latest

# Генерация документации
swag init -g cmd/main.go -o docs
```

---

## Описание эндпоинтов

### Базовый URL

```
http://localhost:8080/api/v1
```

### Аутентификация

Все защищённые эндпоинты требуют JWT токен в заголовке:

```
Authorization: Bearer <your-token>
```

---

### Auth — Аутентификация

| Метод | Endpoint          | Описание                          | Авторизация |
|-------|-------------------|-----------------------------------|-------------|
| POST  | `/auth/register`  | Регистрация нового пользователя   | Не нужна    |
| POST  | `/auth/login`     | Вход в систему                    | Не нужна    |

POST /auth/register

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "Ivan",
  "last_name": "Ivanov"
}
```

**Response (201):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "first_name": "Ivan",
  "last_name": "Ivanov",
  "role": "user",
  "created_at": "2025-01-14T12:00:00Z"
}
```
POST /auth/login

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "Ivan",
    "last_name": "Ivanov",
    "role": "user"
  }
}
```

---

### Users — Пользователи

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| GET | `/users/me` | Получить свой профиль | Нужна |
| PUT | `/users/me` | Обновить профиль | Нужна |
| PUT | `/users/me/password` | Сменить пароль | Нужна |

GET /users/me

**Response (200):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "first_name": "Ivan",
  "last_name": "Ivanov",
  "role": "user",
  "created_at": "2025-01-14T12:00:00Z"
}
```
PUT /users/me

**Request:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith"
}
```

**Response (200):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "first_name": "Jane",
  "last_name": "Smith",
  "role": "user"
}
```
PUT /users/me/password

**Request:**
```json
{
  "old_password": "password123",
  "new_password": "newpassword456"
}
```

**Response (200):**
```json
{
  "message": "Password updated successfully"
}
```

---

### Products — Товары

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| GET | `/products` | Список товаров (с фильтрацией) | Не нужна |
| GET | `/products/{id}` | Получить товар | Не нужна |
| POST | `/products` | Создать товар | Нужна |
| GET | `/products/my` | Мои товары | Нужна |
| PUT | `/products/{id}` | Обновить товар | Нужна |
| DELETE | `/products/{id}` | Удалить товар | Нужна |

GET /products

**Query параметры:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `category` | string | Фильтр по категории |
| `min_price` | float | Минимальная цена |
| `max_price` | float | Максимальная цена |
| `search` | string | Поиск по названию |
| `page` | int | Номер страницы (default: 1) |
| `limit` | int | Количество на странице (default: 10) |
| `sort` | string | Сортировка: price_asc, price_desc, date_asc, date_desc |

**Пример:** `GET /products?category=electronics&min_price=100&max_price=1000&page=1&limit=20`

**Response (200):**
```json
{
  "items": [
    {
      "id": 1,
      "title": "iPhone 15",
      "description": "Новый iPhone",
      "price": 999.99,
      "quantity": 10,
      "category": "electronics",
      "status": "approved",
      "seller": {
        "id": 1,
        "first_name": "Ivan",
        "last_name": "Ivanov"
      }
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```
POST /products

**Request:**
```json
{
  "title": "iPhone 15",
  "description": "Новый iPhone 15 Pro Max",
  "price": 999.99,
  "quantity": 10,
  "category": "electronics"
}
```

**Response (201):**
```json
{
  "id": 1,
  "title": "iPhone 15",
  "description": "Новый iPhone 15 Pro Max",
  "price": 999.99,
  "quantity": 10,
  "category": "electronics",
  "status": "pending",
  "created_at": "2025-01-14T12:00:00Z"
}
```
Товар создаётся со статусом `pending` и требует модерации

---

### Moderation — Модерация

Требуется роль: `moderator` или `admin`

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| GET | `/moderation/products/pending` | Товары на модерации | Нужна (админ или модератор) |
| PUT | `/moderation/products/{id}` | Одобрить/отклонить товар | Нужна (админ или модератор) |

PUT /moderation/products/{id}

**Request:**
```json
{
  "status": "approved",
  "reason": "Товар соответствует правилам"
}
```

Допустимые статусы: `approved`, `rejected`

**Response (200):**
```json
{
  "id": 1,
  "status": "approved",
  "updated_at": "2025-01-14T12:00:00Z"
}
```

---

### Cart — Корзина

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| GET | `/cart` | Получить корзину | Нужна |
| POST | `/cart` | Добавить товар | Нужна |
| PUT | `/cart/{id}` | Обновить количество | Нужна |
| DELETE | `/cart/{id}` | Удалить товар | Нужна |
| DELETE | `/cart/clear` | Очистить корзину | Нужна |

GET /cart

**Response (200):**
```json
{
  "items": [
    {
      "id": 1,
      "product": {
        "id": 1,
        "title": "iPhone 15",
        "price": 999.99
      },
      "quantity": 2,
      "subtotal": 1999.98
    }
  ],
  "total": 1999.98
}
```
POST /cart

**Request:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

**Response (201):**
```json
{
  "id": 1,
  "product_id": 1,
  "quantity": 2
}
```

---

### Orders — Заказы

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| POST | `/orders` | Создать заказ | Нужна |
| POST | `/orders/from-cart` | Заказ из корзины | Нужна |
| GET | `/orders/my` | Мои заказы | Нужна |
| GET | `/orders/{id}` | Получить заказ | Нужна |

POST /orders/from-cart

**Response (201):**
```json
{
  "id": 1,
  "status": "pending",
  "total_amount": 1999.98,
  "payment_status": "pending",
  "items": [
    {
      "product_id": 1,
      "title": "iPhone 15",
      "quantity": 2,
      "price": 999.99
    }
  ],
  "created_at": "2025-01-14T12:00:00Z"
}
```
GET /orders/my

**Response (200):**
```json
{
  "items": [
    {
      "id": 1,
      "status": "paid",
      "total_amount": 1999.98,
      "payment_status": "success",
      "created_at": "2025-01-14T12:00:00Z"
    }
  ],
  "total": 5
}
```

---

### Payments — Оплата

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| POST | `/payments` | Оплатить заказ | Нужна |

POST /payments

Это stub-реализация для демонстрации

**Request:**
```json
{
  "order_id": 1,
  "payment_method": "card"
}
```

Доступные методы: `card`, `wallet`, `cash`

**Response (200):**
```json
{
  "payment_id": "pay_abc123",
  "order_id": 1,
  "status": "success",
  "amount": 1999.98
}
```

---

### Admin — Администрирование

Требуется роль: `admin`

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|------|
| GET | `/admin/users` | Список пользователей | Нужна (админ) |
| GET | `/admin/users/{id}` | Получить пользователя | Нужна (админ) |
| PUT | `/admin/users/{id}/role` | Изменить роль | Нужна (админ) |
| DELETE | `/admin/users/{id}` | Удалить пользователя | Нужна (админ) |
| GET | `/admin/orders` | Все заказы | Нужна (админ) |
| PUT | `/admin/orders/{id}/status` | Изменить статус | Нужна (админ) |

PUT /admin/users/{id}/role

**Request:**
```json
{
  "role": "moderator"
}
```

Доступные роли: `user`, `moderator`, `admin`

**Response (200):**
```json
{
  "id": 2,
  "email": "user@example.com",
  "role": "moderator"
}
```

---

### Коды ошибок

| Код | Описание |
|-----|----------|
| 400 | Bad Request — неверные данные запроса |
| 401 | Unauthorized — требуется авторизация |
| 403 | Forbidden — нет прав доступа |
| 404 | Not Found — ресурс не найден |
| 409 | Conflict — конфликт (например, email занят) |
| 500 | Internal Server Error — ошибка сервера |

### Формат ошибки

```json
{
  "error": "Описание ошибки",
  "code": "ERROR_CODE"
}
```

---

## Тестирование

### Запуск всех тестов

```bash
go test ./... -v
```

### Запуск тестов по слоям

```bash
# Repository тесты
go test ./internal/repository/... -v

# Service тесты
go test ./internal/service/... -v

# Handler тесты
go test ./internal/app/handler/... -v
```

---

## Статусы

### Товары (Product Status)

| Статус | Описание |
|--------|----------|
| `pending` | На модерации |
| `approved` | Одобрен, виден покупателям |
| `rejected` | Отклонён модератором |

### Заказы (Order Status)

| Статус | Описание |
|--------|----------|
| `pending` | Ожидает оплаты |
| `confirmed` | Подтверждён |
| `paid` | Оплачен |
| `shipped` | Отправлен |
| `delivered` | Доставлен |
| `cancelled` | Отменён |

### Оплата (Payment Status)

| Статус | Описание |
|--------|----------|
| `pending` | Ожидает оплаты |
| `success` | Успешно оплачен |
| `failed` | Ошибка оплаты |
| `refunded` | Возврат средств |
