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
Для запуска всей системы выполните следующую команду
```
go run cmd/main.go
```

## Использование

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

### Получение задачи для выполнения
#### Эндпоинт
```
GET /internal/task
```
#### Ответы
##### Успешно получена задача (HTTP 200)
```json
{
    "task":
        {
            "id": <идентификатор задачи>,
            "arg1": <имя первого аргумента>,
            "arg2": <имя второго аргумента>,
            "operation": <операция>,
            "operation_time": <время выполнения операции>
        }
}
```
##### Нет задачи (HTTP 404)
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


### Прием результата обработки данных
#### Эндпоинт
```
POST /internal/task
```
#### Ответы
##### Успешно записан результат (HTTP 200)
```json
{
  "status": "OK"
}
```
##### Нет такой задачи (HTTP 404)
```json
{
  "error": "Not Found"
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

## Примеры использования

### Добавление вычисления арифметического выражения
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/calculate' \
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
#### Невалидные данные
```bash
curl --location 'localhost:8080/api/v1/calculate' \
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
curl --location 'localhost:8080/api/v1/expressions'
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


### Получение выражения по его идентификатору
#### Успешный ответ
```bash
curl --location 'localhost:8080/api/v1/expressions/0'
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
#### Нет такого выражения
```bash
curl --location 'localhost:8080/api/v1/expressions/0'
```
Ожидаемый ответ:
```json
{
  "error": "Not Found"
}
```

### Получение задачи для выполнения
#### Успешный ответ
```bash
curl --location 'localhost:8080/internal/task'
```
Пример ответа:
```json
{
    "task":
        {
            "id": 0,
            "arg1": 2,
            "arg2": 2,
            "operation": "*",
            "operation_time": 1000
        }
}
```
#### Нет задачи
```bash
curl --location 'localhost:8080/internal/task'
```
Пример ответа:
```json
{
  "error": "Not Found"
}
```


### Прием результата обработки данных
#### Успешный ответ
```bash
curl -X POST --location 'localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
  "id": 0,
  "result": 4
}'
```
Пример ответа:
```json
{
  "status": "OK"
}
```
#### Нет такой задачи
```bash
curl -X POST --location 'localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
  "id": 200000,
  "result": 4
}'
```
Пример ответа:
```json
{
  "error": "Not Found"
}
```
#### Невалидные данные
```bash
curl -X POST --location 'localhost:8080/internal/task'
```
Ожидаемый ответ:
```json
{
  "error": "Unprocessable Entity"
}
```
