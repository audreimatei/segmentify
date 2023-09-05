
# segmentify

Dynamic user segmentation service.

## How to run
Clone the repository.

Run the project in Docker:
```
$ docker-compose up --build
```

## How to use
Once launched, the service is available for requests at http://localhost:8080.
At http://localhost:8080/swagger/index.html you can find interactive API docs by SwaggerUI. If you want to play around with the API, you can send a request from interactive docs or use other tools like curl, httpie, Postman, etc.

## Overview of routes
| Task | Method | Route |
| --- | --- | --- |
|Creating a segment | POST | /segments |
|Deleting a segment | DELETE | /segments/{slug} |
|Getting a segment | GET | /segments/{slug} |
|Creating a user | POST | /users |
|Downloading user segments history | GET | /users/{id}/download-segments-history |
|Getting user segments | GET | /users/{id}/segments |
|Updating user segments | PATCH | /users/{id}/segments |

## Dependencies
- [chi](https://github.com/go-chi/chi) lightweight, idiomatic and composable router for building Go HTTP services.
- [pgx](https://github.com/jackc/pgx) pure Go driver and toolkit for PostgreSQL.
- [validator](https://github.com/go-playground/validator) Go Struct and Field validation.
- [swag](https://github.com/swaggo/swag) automatically generate RESTful API documentation with Swagger 2.0 for Go.