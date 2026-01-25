# Руководство по использованию API (API Usage Cases)

## 1. Сокращение ссылки (Save URL)

Создает короткий алиас для длинного URL.
* **Метод:** `POST`
* **Эндпоинт:** `/url`
* **Требует авторизации:** Да

**Пример запроса (с автоматическим алиасом):**
```bash
curl -X POST http://localhost:8080/url \
    --user admin:password \
    -H "Content-Type: application/json" \
    -d '{"url": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Access-Control-Allow-Credentials"}'
```

**Ответ (Успех - 201 Created):**
```json
{
  "status": "OK",
  "alias": "AAMB6U"
}
```

**Пример запроса (с кастомным алиасом):**
```bash
curl -X POST http://localhost:8080/url \
    --user admin:password \
    -H "Content-Type: application/json" \
    -d '{"url": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Access-Control-Allow-Credentials", "alias": "some-alias"}'
```

**Ответ (Успех - 201 Created):**
```json
{
  "status": "OK",
  "alias": "some-alias"
}
```

**Возможные ошибки**

### 400 Bad Request
**Описание:** Невалидный URL или пустое тело запроса.

### 409 Conflict
**Описание:** Такой алиас уже занят.


## 2. Переход по ссылке (Redirect)

Перенаправляет пользователя на исходный URL.
* **Метод:** `GET`
* **Эндпоинт:** `/{alias}`
* **Требует авторизации:** Нет

```bash
curl -v http://localhost:8080/some-alias
```

**Ответ (Успех - 302 Found):**
В заголовках ответа вы увидите:

```plaintext
< HTTP/1.1 302 Found
< Location: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Access-Control-Allow-Credentials
```

**Возможные ошибки**
### 404 Not Found
**Описание:** Ссылка с таким алиасом не найдена.


## 3. Удаление ссылки (Delete URL)
Удаляет ссылку из базы данных.
* **Метод:** `DELETE`
* **Эндпоинт:** `/url/{alias}`
* **Требует авторизации:** Да

**Пример запроса:**
```bash
curl -X DELETE http://localhost:8080/url/some-alias \
    --user admin:password
```

**Возможные ошибки**
### 404 Not Found
**Описание:** Ссылка, которую вы пытаетесь удалить, не существует.