package update

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	SegmentsToAdd    []models.SegmentToAdd    `json:"segments_to_add" validate:"required"`
	SegmentsToRemove []models.SegmentToRemove `json:"segments_to_remove" validate:"required"`
}

type UserSegmentsUpdater interface {
	UpdateUserSegments(
		ctx context.Context,
		id int64,
		segmentsToAdd []models.SegmentToAdd,
		segmentsToRemove []models.SegmentToRemove,
	) error
}

// @Summary	Updating user segments
// @Tags		users
// @Param		id		path	string	true	"User ID"
// @Param		body	body	Request	true	"Segments to add/remove"
// @Success	204
// @Failure	400	{object}	resp.ErrResponse
// @Failure	404	{object}	resp.ErrResponse
// @Failure	422	{object}	resp.ErrResponse
// @Failure	500	{object}	resp.ErrResponse
// @Router		/users/{id}/segments [patch]
func New(ctx context.Context, log *slog.Logger, userSegmentsUpdater UserSegmentsUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.update.New"

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

		var req Request

		err = render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Info("request body is empty")

			render.Render(w, r, resp.ErrInvalidRequest("request body is empty"))
			return
		}
		if err != nil {
			log.Info("failed to decode request body", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Info("invalid request", sl.Err(err))

			render.Render(w, r, resp.ValidationError(validateErr))
			return
		}

		if len(req.SegmentsToAdd) > 0 &&
			len(req.SegmentsToRemove) > 0 &&
			checkSegmentOverlap(req.SegmentsToAdd, req.SegmentsToRemove) {
			log.Info("segmentsToAdd and segmentsToRemove overlap")

			render.Render(w, r, resp.ErrInvalidRequest("segments_to_add and segments_to_remove overlap"))
			return
		}

		if err = userSegmentsUpdater.UpdateUserSegments(
			ctx,
			id,
			req.SegmentsToAdd,
			req.SegmentsToRemove,
		); err != nil {
			var errUserSegmentExists *storage.ErrUserSegmentExists
			var errUserNotFound *storage.ErrUserNotFound
			var errSegmentNotFound *storage.ErrSegmentNotFound
			var errUserSegmentNotFound *storage.ErrUserSegmentNotFound

			if errors.As(err, &errUserSegmentExists) {
				render.Render(w, r, resp.ErrInvalidRequest(errUserSegmentExists.Error()))
				return
			}
			if errors.As(err, &errUserNotFound) {
				render.Render(w, r, resp.ErrNotFound(errUserNotFound.Error()))
				return
			}
			if errors.As(err, &errSegmentNotFound) {
				render.Render(w, r, resp.ErrNotFound(errSegmentNotFound.Error()))
				return
			}
			if errors.As(err, &errUserSegmentNotFound) {
				render.Render(w, r, resp.ErrNotFound(errUserSegmentNotFound.Error()))
				return
			}
			log.Error("failed to update user segments", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to update user segments"))
			return
		}

		log.Info("user segments updated")

		w.WriteHeader(http.StatusNoContent)
	}
}

func checkSegmentOverlap(segmentsToAdd []models.SegmentToAdd, segmentsToRemove []models.SegmentToRemove) bool {
	for _, s1 := range segmentsToAdd {
		for _, s2 := range segmentsToRemove {
			if s1.Slug == s2.Slug {
				return true
			}
		}
	}
	return false
}
