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
			render.Render(w, r, resp.ErrInvalidRequest("user id is invalid"))
			return
		}

		segments, err := userSegmentsGetter.GetUserSegments(ctx, id)
		if err != nil {
			var errUserNotFound *storage.ErrUserNotFound
			var errUserSegmentNotFound *storage.ErrUserSegmentNotFound

			if errors.As(err, &errUserNotFound) {
				render.Render(w, r, resp.ErrNotFound(errUserNotFound.Error()))
				return
			} else if errors.As(err, &errUserSegmentNotFound) {
				render.Render(w, r, resp.ErrNotFound(errUserSegmentNotFound.Error()))
				return
			}
			log.Error("failed to get user segment", sl.Err(err))
			render.Render(w, r, resp.ErrInternal("failed to get user segment"))
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{ID: id, Segments: segments})
	}
}
