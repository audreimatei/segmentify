
# segmentify

Dynamic user segmentation service.

## How to run
Clone the repository. Rename *.env.example* to *.env* and fill in the environment variables.

To run the project in Docker:
```
$ docker-compose up
```

## How to use
- `POST /segments ` create segment
- `GET /segments/{slug} ` get segment
- `DELETE /segments ` delete segment

- `POST /users ` create user
- `GET /users/{userID}/segments ` get active user segments
- `PATCH /users/{userID}/segments ` add user to segment


## Dependencies
- [chi](github.com/go-chi/chi) lightweight, idiomatic and composable router for building Go HTTP services.
- [pgx](https://github.com/jackc/pgx) pure Go driver and toolkit for PostgreSQL.
- [validator](github.com/go-playground/validator) Go Struct and Field validation.