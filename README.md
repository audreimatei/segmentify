# segmentify

Сервис динамического сегментирования пользователей.

## Какую проблему решает
Проходят различные продуктовые эксперименты — тесты новой функциональности, тесты интерфейса, скидочные и т.д. Хотелось бы иметь сервис, который хранит сегменты пользователей и предоставляет удобный интерфейс работы с ними.

## Как запустить
Клонируйте репозиторий.

Запустите проект в Docker:
```
$ docker-compose up --build
```

## Как пользоваться
После запуска сервис доступен для запросов по адресу http://localhost:8080.
А по адресу http://localhost:8080/swagger/index.html находится интерактивная документация по API. Вы можете отправить запрос из интерактивной документации или воспользоваться curl, httpie, Postman и т.д.

## Обзор эндпоинтов
| Задача | Метод | Эндпоинт |
| --- | --- | --- |
| Создание сегмента | POST | /segments |
| Удаление сегмента | DELETE | /segments/{slug} |
| Получение сегмента | GET | /segments/{slug} |
| Создание пользователя | POST | /users |
| Выгрузка истории пользовательских сегментов | GET | /users/{id}/download-segments-history |
| Получение сегментов пользователя | GET | /users/{id}/segments |
| Обновление сегментов пользователя | PATCH | /users/{id}/segments |

## Особенности реализации дополнительных заданий
- **Первое задание**. При добавлении/удалении сегмента у пользователя, создаётся запись в users_segments_history.

- **Второе задание**. В БД к таблице users_segments добавил поле expire_at — дата и время по которое пользователь должен находится в сегменте. При получении сегментов пользователя проводим фильтрацию по полю exipre_at, чтобы не получать истёкшие записи. Горутина startSheduler каждый час вызывает функцию RemoveExpiredUsersSegments и удаляет все истёкшие записи из users_segments.

- **Третье задание**. В БД к таблице segments добавил percent — процент пользователей, которые будут попадать в сегмент автоматически. Если при создании сегмента передаётся percent != 0, то выбираем случайных (ORDER BY RANDOM()) пользователей нужного количества и создаём записи в users_segments и users_segments_history c помощью PostgreSQL COPY протокола.

## Как запустить интеграционные тесты
Запустите тестовую среду в Docker:
```
$ docker-compose --file docker-compose.test.yml up --build
```
Запустите тесты:
```
$ go test -v -count 1 -timeout 30s ./test/integration
```

## Зависимости проекта
- [chi](https://github.com/go-chi/chi) lightweight, idiomatic and composable router for building Go HTTP services.
- [pgx](https://github.com/jackc/pgx) pure Go driver and toolkit for PostgreSQL.
- [validator](https://github.com/go-playground/validator) Go Struct and Field validation.
- [swag](https://github.com/swaggo/swag) automatically generate RESTful API documentation with Swagger 2.0 for Go.