package create

import (
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	ID int64 `json:"id"`
}

type UserCreator interface {
	CreateUser() (int64, error)
}

//	@Summary	Creating a user
//	@Tags		users
//	@Success	201	{object}	Response
//	@Failure	500	{object}	resp.ErrResponse
//	@Router		/users [post]
func New(log *slog.Logger, userCreator UserCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.create.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, err := userCreator.CreateUser()
		if err != nil {
			log.Error("failed to create user", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to create user"))
			return
		}

		log.Info("user created", slog.Int64("id", userID))

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{ID: userID})
	}
}
