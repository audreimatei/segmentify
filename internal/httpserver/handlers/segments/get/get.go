package getbyslug

import (
	"errors"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	models.Segment
}

type SegmentGetter interface {
	GetSegment(slug string) (models.Segment, error)
}

// @Summary	Getting a segment
// @Tags		segments
// @Param		slug	path		string	true "Segment slug"
// @Success	200		{object} Response
// @Failure	400		{object}	resp.ErrResponse
// @Failure	404		{object}	resp.ErrResponse
// @Failure	500		{object}	resp.ErrResponse
// @Router		/segments/{slug} [get]
func New(log *slog.Logger, segmentGetter SegmentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		slug := chi.URLParam(r, "slug")
		if slug == "" {
			log.Info("slug is invalid")

			render.Render(w, r, resp.ErrInvalidRequest("slug is invalid"))
			return
		}

		log.Info("slug extracted from path", slog.String("slug", slug))

		segment, err := segmentGetter.GetSegment(slug)
		if err != nil {
			if errors.Is(err, storage.ErrSegmentNotFound) {
				log.Info("segment not found", slog.String("slug", slug))

				render.Render(w, r, resp.ErrNotFound("segment not found"))
				return
			}
			log.Error("failed to get segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to get segment"))
			return
		}

		log.Info(
			"segment received",
			slog.Int64("id", segment.ID),
			slog.String("slug", segment.Slug),
		)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{Segment: segment})
	}
}