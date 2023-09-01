package update

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	SegmentsToAdd    []string `json:"segments_to_add" validate:"required"`
	SegmentsToRemove []string `json:"segments_to_remove" validate:"required"`
}

func checkSegmentOverlap(segmentsToAdd []string, segmentsToRemove []string) bool {
	for _, s1 := range segmentsToAdd {
		for _, s2 := range segmentsToRemove {
			if s1 == s2 {
				return true
			}
		}
	}
	return false
}

type UserSegmentsUpdater interface {
	UpdateUserSegments(id int64, segmentsToAdd []string, segmentsToRemove []string) error
}

func New(log *slog.Logger, userSegmentsUpdater UserSegmentsUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			log.Info("userID is invalid", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("user_id is invalid"))
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

		err = userSegmentsUpdater.UpdateUserSegments(
			userID,
			req.SegmentsToAdd,
			req.SegmentsToRemove,
		)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found", slog.Int64("id", userID))

				render.Render(w, r, resp.ErrNotFound("user not found"))
				return
			} else if errors.Is(err, storage.ErrSegmentNotFound) {
				log.Info("segment not found", sl.Err(err))

				render.Render(w, r, resp.ErrNotFound("segment not found"))
				return
			} else if errors.Is(err, storage.ErrUserSegmentExists) {
				log.Info("user segment exists", sl.Err(err))

				render.Render(w, r, resp.ErrInvalidRequest("user segment exists"))
				return
			} else if errors.Is(err, storage.ErrUserSegmentNotFound) {
				log.Info("user segment not found", sl.Err(err))

				render.Render(w, r, resp.ErrNotFound("user segment not found"))
				return
			}
			log.Error("failed to update user segments", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to update user segments"))
			return
		}

		log.Info("user segments updated")

		render.Status(r, http.StatusNoContent)
	}
}
