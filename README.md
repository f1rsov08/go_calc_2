# Распределённый вычислитель арифметических выражений

## Описание проекта

Данный проект представляет собой распределённую систему для вычисления арифметических выражений, состоящую из двух основных компонентов: оркестратора и агента. Оркестратор принимает арифметические выражения, разбивает их на задачи и управляет их выполнением. Агент получает задачи от оркестратора, выполняет вычисления и возвращает результаты.

## Структура проекта

* Оркестратор: Сервер, который предоставляет API для добавления выражений, получения статусов вычислений и управления задачами.

* Агент: Демон, который выполняет вычисления и взаимодействует с оркестратором.

![мавмва](https://github.com/user-attachments/assets/31650fb9-a3d7-43dc-a273-7c7784f77618)
## Установка

1. Склонируйте репозиторий
```
git clone https://github.com/f1rsov08/go_calc_2
cd go_calc_2
```
2. Установите необходимые зависимости
```
go get -d github.com/joho/godotenv/cmd/godotenv
```
3. Создайте файл .env в корне проекта и добавьте следующие переменные среды
```
TIME_ADDITION_MS=<время_выполнения_сложения>
TIME_SUBTRACTION_MS=<время_выполнения_вычитания>
TIME_MULTIPLICATIONS_MS=<время_выполнения_умножения>
TIME_DIVISIONS_MS=<время_выполнения_деления>
COMPUTING_POWER=<количество_горутин>
WAIT_TIME=<периодичность отправки запросов агента оркестратору>
ORCHESTRATOR_PORT=<порт оркестратора>
```

## Запуск
Для запуска системы выполните следующую команду
```
go run cmd/main.go
```

## Использование

### Регистрация
#### Эндпоинт
```
POST /api/v1/register
```
#### Запрос
```json
{
  "login": <логин>,
  "password": <пароль>
}
```
#### Ответы
##### Регистрация прошла успешно (HTTP 200)
```json
{
  "status": "OK"
}
```
##### Невалидные данные (HTTP 422)
```json
{
  "error": "Unprocessable Entity"
}
```
##### Что-то пошло не так (HTTP 500)
```json
{
  "error": "Internal Server Error"
}
```

### Вход
#### Эндпоинт
```
POST /api/v1/login
```
#### Запрос
```json
{
  "login": <логин>,
  "password": <пароль>
}
```
#### Ответы
##### Успешный вход (HTTP 200)
```json
{
  "token": <JWT токен для последующей авторизации>
}
```
##### Неверный логин или пароль (HTTP 401)
```json
{
  "error": "Unauthorized"
}
```
##### Невалидные данные (HTTP 422)
```json
{
  "error": "Unprocessable Entity"
}
```
##### Что-то пошло не так (HTTP 500)
```json
{
  "error": "Internal Server Error"
}
```

> [!WARNING]  
> ## У всех последующих запросов в заголовке должен быть JWT-токен! ```Authorization: Bearer <токен>```
### Добавление вычисления арифметического выражения
#### Эндпоинт
```
POST /api/v1/calculate
```
#### Запрос
```json
{
  "expression": <строка с выражение>
}
```
#### Ответы
##### Выражение принято для вычисления (HTTP 201)
```json
{
  "id": <уникальный идентификатор выражения>
}
```
##### Неверный токен (HTTP 401)
```json
{
  "error": "Unauthorized"
}
```
##### Невалидные данные (HTTP 422)
```json
{
  "error": "Unprocessable Entity"
}
```
##### Что-то пошло не так (HTTP 500)
```json
{
  "error": "Internal Server Error"
}
```


### Получение списка выражений
#### Эндпоинт
```
GET /api/v1/expressions
```
#### Ответы
##### Успешно получен список выражений (HTTP 200)
```json
{
    "expressions": [
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        },
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
    ]
}
```
Статусы вычисления выражения:
* waiting - выражение ждет вычисления
* complete - выражение вычислено
* error - ошибка во время вычисления
##### Неверный токен (HTTP 401)
```json
{
  "error": "Unauthorized"
}
```
##### Что-то пошло не так (HTTP 500)
```json
{
  "error": "Internal Server Error"
}
```

### Получение выражения по его идентификатору
#### Эндпоинт
```
GET /api/v1/expressions/:id
```
#### Ответы
##### Успешно получено выражение (HTTP 200)
```json
{
    "expression":
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
}
```
##### Доступ к выражению запрещен (HTTP 403)
```json
{
  "error": "Forbidden"
}
```
##### Нет такого выражения (HTTP 404)
```json
{
  "error": "Not Found"
}
```
##### Что-то пошло не так (HTTP 500)
```json
{
  "error": "Internal Server Error"
}
```

## Примеры использования
### Регистрация
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "ivanivanov",
  "password": "12345678"
}'
```
Ответ:
```json
{
  "status": "OK"
}
```
#### Невалидные данные
```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "ivanivanov"
}'
```
Ответ:
```json
{
  "error": "Unprocessable Entity"
}
```
### Вход
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "ivanivanov",
  "password": "12345678"
}'
```
Ответ:
```json
{
  "token": <токен>
}
```
#### Неверный логин или пароль
```bash
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "какой-то несуществующий логин",
  "password": "какой-то неверный пароль"
}'
```
Ответ:
```json
{
  "error": "Unauthorized"
}
```
#### Невалидные данные
```bash
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "password": "12345678"
}'
```
Ответ:
```json
{
  "error": "Unprocessable Entity"
}
```
### Добавление вычисления арифметического выражения
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Authorization: Bearer <токен>' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```
Пример ответа:
```json
{
  "id": 0
}
```
#### Неверный токен
```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "5+2"
}'
```
Ожидаемый ответ:
```json
{
  "error": "Unauthorized"
}
```
#### Невалидные данные
```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Authorization: Bearer <токен>' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "42))"
}'
```
Ожидаемый ответ:
```json
{
  "error": "Unprocessable Entity"
}
```


### Получение списка выражений
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer <токен>'
```
Пример ответа:
```json
{
    "expressions": [
        {
            "id": 0,
            "status": "complete",
            "result": 55
        },
        {
            "id": 1,
            "status": "waiting",
            "result": 0
        }
    ]
}
```
#### Неверный токен
```bash
curl --location 'localhost:8080/api/v1/expressions'
```
Ожидаемый ответ:
```json
{
  "error": "Unauthorized"
}
```


### Получение выражения по его идентификатору
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/expressions/0' \
--header 'Authorization: Bearer <токен>'
```
Пример ответа:
```json
{
    "expression":
        {
            "id": 0,
            "status": "complete",
            "result": 55
        }
}
```
#### Доступ к выражению запрещен
```bash
curl --location 'localhost:8080/api/v1/expressions/0' \
--header 'Authorization: Bearer <неверный токен>'
```
Ожидаемый ответ:
```json
{
  "error": "Forbidden"
}
```
#### Нет такого выражения
```bash
curl --location 'localhost:8080/api/v1/expressions/1000' \
--header 'Authorization: Bearer <токен>'
```
Ожидаемый ответ:
```json
{
  "error": "Not Found"
}
```
