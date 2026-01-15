# Финальный артефакт

## Таблица соответствия

### 1 Аутентификация и авторизация

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-1** | Регистрация пользователя | `POST /api/v1/auth/register` | **Given** валидные email/password (≥6 символов) — **When** POST /auth/register — **Then** 201 Created, возвращается user без пароля | Пароль ≥ 6 символов; email уникален; новый пользователь получает роль "user" | `TestUserService_Register_Success`; `TestUserService_Register_SetsDefaultRole` | 400 слабый пароль (`TestUserService_Register_WeakPassword`); 409 email занят (`TestUserService_Register_UserAlreadyExists`) |
| **ST-2** | Авторизация | `POST /api/v1/auth/login` | **Given** зарегистрированный пользователь — **When** POST /auth/login с email/password — **Then** 200 OK, возвращается JWT токен | Пароль сверяется через bcrypt; токен содержит user_id, email, role | `TestUserService_Login_Success` | 401 неверный email/password (`TestUserService_Login_InvalidPassword`; `TestUserService_Login_UserNotFound`) |

### 2 Управление пользователями

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-3** | Получение своего профиля | `GET /api/v1/users/me` | **Given** авторизован — **When** GET /users/me — **Then** 200 OK, данные текущего пользователя | Токен валиден; пользователь существует | `TestUserHandler_GetMe_Success` | 401 без токена; 401 невалидный токен |
| **ST-4** | Получение пользователя по ID | `GET /api/v1/users/{id}` | **Given** авторизован как admin или запрашивает свой профиль — **When** GET /users/{id} — **Then** 200 OK, данные пользователя | Пользователь может видеть только свой профиль; admin видит любой | `TestUserService_GetByID_Success` | 403 чужой профиль не-админом; 404 не найден (`TestUserService_GetByID_NotFound`) |
| **ST-5** | Список пользователей | `GET /api/v1/users` | **Given** роль admin — **When** GET /users?page=1&page_size=20 — **Then** 200 OK, список пользователей с пагинацией | Только admin может видеть список; пагинация | `TestUserService_GetAll_Success` | 403 не admin; 401 без токена |
| **ST-6** | Обновление профиля | `PUT /api/v1/users/{id}` | **Given** авторизован как владелец или admin — **When** PUT /users/{id} с first_name, last_name — **Then** 200 OK, профиль обновлён | Можно обновить только свой профиль; admin может любой | `TestUserService_Update_Success` | 403 чужой профиль; 404 не найден |
| **ST-7** | Изменение пароля | `PUT /api/v1/users/{id}/password` | **Given** авторизован как владелец — **When** PUT /users/{id}/password с old_password, new_password — **Then** 200 OK, пароль изменён | Старый пароль должен совпадать; новый ≥ 6 символов | `TestUserService_ChangePassword_Success` | 400 неверный старый пароль (`TestUserService_ChangePassword_WrongOldPassword`); 400 слабый новый пароль |
| **ST-8** | Изменение роли | `PUT /api/v1/users/{id}/role` | **Given** роль admin — **When** PUT /users/{id}/role с новой ролью — **Then** 200 OK, роль изменена | Только admin может менять роли; допустимые роли: user, moderator, admin | `TestUserService_UpdateRole_Success` | 403 не admin (`TestUserService_UpdateRole_Forbidden`); 404 пользователь не найден |
| **ST-9** | Удаление пользователя | `DELETE /api/v1/users/{id}` | **Given** роль admin — **When** DELETE /users/{id} — **Then** 204 No Content, пользователь удалён | Только admin может удалять; каскадное удаление связанных данных | `TestUserService_Delete_Success` | 403 не admin; 404 не найден |

### 3 Управление товарами

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-10** | Создание товара | `POST /api/v1/products` | **Given** авторизован как user — **When** POST /products с title, price > 0 — **Then** 201 Created, товар со статусом "pending" | Цена > 0; название 1-255 символов; новый товар в статусе "pending" | `TestProductService_Create_Success` | 400 пустое название (`TestProductService_Create_InvalidTitle_Empty`); 400 цена ≤ 0 (`TestProductService_Create_InvalidPrice_Zero`) |
| **ST-11** | Список товаров | `GET /api/v1/products` | **Given** любой авторизованный — **When** GET /products?page=1&page_size=20 — **Then** 200 OK, список товаров с пагинацией | Пагинация: page ≥ 1, page_size 1-100 (default 20); фильтрация по category, price_min, price_max, status | `TestProductService_GetAll_Success`; `TestProductService_GetAll_DefaultPagination` | 401 без токена |
| **ST-12** | Получение товара по ID | `GET /api/v1/products/{id}` | **Given** авторизован — **When** GET /products/{id} — **Then** 200 OK, данные товара | Товар существует | `TestProductService_GetByID_Success` | 404 не найден (`TestProductService_GetByID_NotFound`) |
| **ST-13** | Обновление товара | `PUT /api/v1/products/{id}` | **Given** авторизован как владелец или moderator/admin — **When** PUT /products/{id} — **Then** 200 OK, товар обновлён | Владелец или moderator/admin; валидация title, price | `TestProductService_Update_Success`; `TestProductService_Update_OwnerCanUpdate`; `TestProductService_Update_ModeratorCanUpdate` | 403 не владелец (`TestProductService_Update_NotOwner`); 404 не найден; 400 невалидные данные |
| **ST-14** | Удаление товара | `DELETE /api/v1/products/{id}` | **Given** авторизован как владелец или admin — **When** DELETE /products/{id} — **Then** 204 No Content, товар удалён | Только владелец или admin может удалять | `TestProductService_Delete_Success` | 403 не владелец (`TestProductService_Delete_NotOwner`); 404 не найден |
| **ST-15** | Модерация товара | `PUT /api/v1/products/{id}/status` | **Given** роль moderator/admin — **When** PUT /products/{id}/status с "approved"/"rejected" — **Then** 200 OK, статус изменён | Только moderator/admin могут менять статус; допустимые значения: pending, approved, rejected | `TestProductService_UpdateStatus_Success` | 403 обычный user; 404 товар не найден |
| **ST-16** | Товары продавца | `GET /api/v1/products/seller/{seller_id}` | **Given** авторизован — **When** GET /products/seller/{seller_id} — **Then** 200 OK, список товаров продавца | Пагинация; фильтрация по seller_id | `TestProductService_GetBySeller_Success` | 401 без токена |

### 4 Корзина

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-17** | Добавление в корзину | `POST /api/v1/cart` | **Given** товар approved и quantity > 0 — **When** POST /cart с product_id и quantity — **Then** 201 Created, товар в корзине | Товар должен быть approved; запрошенное количество ≤ доступного | `TestOrderService_AddToCart_Success` | 400 товар недоступен (`TestOrderService_AddToCart_ProductUnavailable`); 400 недостаточно на складе |
| **ST-18** | Просмотр корзины | `GET /api/v1/cart` | **Given** авторизован — **When** GET /cart — **Then** 200 OK, содержимое корзины с итоговой суммой | Показывает только свою корзину; итоговая сумма | `TestOrderService_GetCart_Success` | 401 без токена |
| **ST-19** | Обновление количества в корзине | `PUT /api/v1/cart/{id}` | **Given** авторизован, товар в корзине — **When** PUT /cart/{id} с quantity — **Then** 200 OK, количество обновлено | Количество ≤ доступного на складе; владелец корзины | `TestOrderService_UpdateCartItem_Success` | 400 недостаточно на складе; 403 чужая корзина; 404 товар не найден |
| **ST-20** | Удаление из корзины | `DELETE /api/v1/cart/{id}` | **Given** авторизован, товар в корзине — **When** DELETE /cart/{id} — **Then** 204 No Content, товар удалён | Только владелец корзины может убирать товары из нее | `TestOrderService_RemoveFromCart_Success` | 403 чужая корзина; 404 не найден |
| **ST-21** | Очистка корзины | `DELETE /api/v1/cart` | **Given** авторизован — **When** DELETE /cart — **Then** 204 No Content, корзина очищена | Удаляются все товары из корзины пользователя | `TestOrderService_ClearCart_Success` | 401 без токена |

### 5 Заказы

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-22** | Оформление заказа | `POST /api/v1/orders` | **Given** корзина не пуста, товары доступны — **When** POST /orders — **Then** 201 Created, заказ создан, корзина очищена | Корзина не пуста; все товары approved; количество ≤ наличия; остаток на складе уменьшается | `TestOrderService_CreateOrder_Success`; `TestOrderService_CreateOrderFromCart_Success` | 400 пустая корзина (`TestOrderService_CreateOrder_EmptyCart`); 400 товар недоступен (`TestOrderService_CreateOrder_ProductUnavailable`); 400 недостаточно на складе (`TestOrderService_CreateOrder_InsufficientStock`) |
| **ST-23** | Мои заказы | `GET /api/v1/orders` | **Given** авторизован — **When** GET /orders?page=1&page_size=20 — **Then** 200 OK, только свои заказы с пагинацией | Пользователь видит только свои заказы; пагинация | `TestOrderService_GetUserOrders_Success` | 401 без токена |
| **ST-24** | Детали заказа | `GET /api/v1/orders/{id}` | **Given** авторизован как владелец или admin — **When** GET /orders/{id} — **Then** 200 OK, детали заказа с items | Владелец заказа или admin | `TestOrderService_GetOrderByID_Success` | 403 чужой заказ (`TestOrderService_GetOrderByID_NotOwner`); 404 не найден (`TestOrderService_GetOrderByID_NotFound`) |
| **ST-25** | Обновление статуса заказа | `PUT /api/v1/orders/{id}/status` | **Given** роль admin — **When** PUT /orders/{id}/status — **Then** 200 OK, статус обновлён | Только admin; допустимые переходы статусов | `TestOrderService_UpdateOrderStatus_Success` | 403 не admin; 404 не найден; 400 недопустимый переход статуса |
| **ST-26** | Оплата заказа | `POST /api/v1/orders/{id}/pay` | **Given** авторизован как владелец, статус pending — **When** POST /orders/{id}/pay — **Then** 200 OK, оплата обработана | Владелец заказа; заказ в статусе pending; валидный способ оплаты | `TestOrderService_ProcessPayment_Success` | 403 чужой заказ; 400 заказ уже оплачен; 404 не найден |
| **ST-27** | Все заказы (admin) | `GET /api/v1/admin/orders` | **Given** роль admin — **When** GET /admin/orders?page=1&page_size=20 — **Then** 200 OK, все заказы системы с пагинацией | Только admin видит все заказы; пагинация | `TestOrderService_GetAllOrders_Success` | 403 не admin |

### 6 Системные

| Story ID | Краткое описание | Эндпоинты | Критерии приёмки (GWT) | Бизнес-правила | Юнит-тесты | Негативные кейсы |
|----------|------------------|-----------|------------------------|----------------|------------|------------------|
| **ST-28** | Проверка работы сервиса | `GET /health` | **Given** сервис запущен — **When** GET /health — **Then** 200 OK, `{"status":"ok"}` | Публичный эндпоинт; не требует авторизации | — | 500 если БД недоступна |
