package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	ID       int64    `json:"id"`
	Segments []string `json:"segments"`
}

type UserSegmentsGetter interface {
	GetUserSegments(ctx context.Context, id int64) ([]string, error)
}

// @Summary	Getting user segments
// @Tags		users
// @Param		id	path		string	true	"User ID"
// @Success	200	{object}	Response
// @Failure	400	{object}	resp.ErrResponse
// @Failure	404	{object}	resp.ErrResponse
// @Failure	500	{object}	resp.ErrResponse
// @Router		/users/{id}/segments [get]
func New(ctx context.Context, log *slog.Logger, userSegmentsGetter UserSegmentsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Info("user id is invalid", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("user id is invalid"))
			return
		}

		segments, err := userSegmentsGetter.GetUserSegments(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found", slog.Int64("id", id))

				render.Render(w, r, resp.ErrNotFound("user not found"))
				return
			} else if errors.Is(err, storage.ErrUserSegmentNotFound) {
				log.Info("user segments not found", slog.Int64("id", id))

				render.Render(w, r, resp.ErrNotFound("user segments not found"))
				return
			}
			log.Error("failed to get user segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to get user segment"))
			return
		}

		log.Info("user segments received")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{ID: id, Segments: segments})
	}
}
