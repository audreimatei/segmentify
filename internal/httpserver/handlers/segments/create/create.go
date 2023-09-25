package create

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@v2.33.1 --name=SegmentCreator
type SegmentCreator interface {
	CreateSegment(ctx context.Context, segment models.Segment) (models.Segment, error)
}

// @Summary	Creating a segment
// @Tags		segments
// @Param		body	body		models.Segment	true	"Segment"
// @Success	201		{object}	models.Segment
// @Failure	400		{object}	resp.ErrResponse
// @Failure	422		{object}	resp.ErrResponse
// @Failure	500		{object}	resp.ErrResponse
// @Router		/segments [post]
func New(ctx context.Context, log *slog.Logger, segmentCreator SegmentCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.create.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req models.Segment

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

		dbSegment, err := segmentCreator.CreateSegment(ctx, req)
		if err != nil {
			var errSegmentExists *storage.ErrSegmentExists
			if errors.As(err, &errSegmentExists) {
				render.Render(w, r, resp.ErrInvalidRequest(errSegmentExists.Error()))
				return
			}
			log.Error("failed to create segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to create segment"))
			return
		}

		log.Info("segment created", slog.String("slug", dbSegment.Slug))

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dbSegment)
	}
}
