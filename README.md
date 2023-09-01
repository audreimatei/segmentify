
# segmentify

Dynamic user segmentation service.

## How to run
Clone the repository.

Run the project in Docker:
```
$ docker-compose up --build
```
Once launched, the service is available at http://localhost:8080.

## Overview of routes
| Task | Method | Route |
| --- | --- | --- |
|Creating a segment | POST | /segments |
|Getting a segment by slug | GET | /segments/{slug} |
|Deleting a segment | DELETE | /segments |
|Creating a user | POST | /users |
|Getting user segments | GET | /users/{userID}/segments |
|Downloading user segments history | GET | /users/{userID}/download-segments-history |
|Updating user segments | PATCH | /users/{userID}/segments |

## Dependencies
- [chi](github.com/go-chi/chi) lightweight, idiomatic and composable router for building Go HTTP services.
- [pgx](https://github.com/jackc/pgx) pure Go driver and toolkit for PostgreSQL.
- [validator](github.com/go-playground/validator) Go Struct and Field validation.