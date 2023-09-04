package create

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Slug string `json:"slug" validate:"required"`
}

type Response struct {
	models.Segment
}

//go:generate go run github.com/vektra/mockery/v2@v2.33.1 --name=SegmentCreator
type SegmentCreator interface {
	CreateSegment(slug string) (models.Segment, error)
}

// @Summary	Creating a segment
// @Tags		segments
// @Param		body	body		Request	true	"Segment slug"
// @Success	201		{object}	Response
// @Failure	400		{object}	resp.ErrResponse
// @Failure	422		{object}	resp.ErrResponse
// @Failure	500		{object}	resp.ErrResponse
// @Router		/segments [post]
func New(log *slog.Logger, segmentCreator SegmentCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.create.New"

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

		segment, err := segmentCreator.CreateSegment(req.Slug)
		if err != nil {
			if errors.Is(err, storage.ErrSegmentExists) {
				log.Info("segment already exists", slog.String("slug", req.Slug))

				render.Render(w, r, resp.ErrInvalidRequest("segment already exists"))
				return
			}
			log.Error("failed to create segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to create segment"))
			return
		}

		log.Info(
			"segment created",
			slog.Int64("id", segment.ID),
			slog.String("slug", segment.Slug),
		)

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{Segment: segment})
	}
}
