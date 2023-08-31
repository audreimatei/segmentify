package delete

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Slug string `json:"slug" validate:"required"`
}

type SegmentDeleter interface {
	DeleteSegment(slug string) error
}

func New(log *slog.Logger, segmentDeleter SegmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
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

		err = segmentDeleter.DeleteSegment(req.Slug)
		if err != nil {
			if errors.Is(err, storage.ErrSegmentNotFound) {
				log.Info("segment not found", slog.String("slug", req.Slug))

				render.Render(w, r, resp.ErrNotFound("segment not found"))
				return
			}
			log.Error("failed to delete segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to delete segment"))
			return
		}

		log.Info("segment deleted", slog.String("slug", req.Slug))
	}
}
