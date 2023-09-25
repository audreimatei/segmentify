package getbyslug

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/models"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SegmentGetter interface {
	GetSegment(ctx context.Context, slug string) (models.Segment, error)
}

// @Summary	Getting a segment
// @Tags		segments
// @Param		slug	path		string	true "Segment slug"
// @Success	200		{object} models.Segment
// @Failure	400		{object}	resp.ErrResponse
// @Failure	404		{object}	resp.ErrResponse
// @Failure	500		{object}	resp.ErrResponse
// @Router		/segments/{slug} [get]
func New(ctx context.Context, log *slog.Logger, segmentGetter SegmentGetter) http.HandlerFunc {
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

		dbSegment, err := segmentGetter.GetSegment(ctx, slug)
		if err != nil {
			var errSegmentNotFound *storage.ErrSegmentNotFound

			if errors.As(err, &errSegmentNotFound) {
				render.Render(w, r, resp.ErrNotFound(errSegmentNotFound.Error()))
				return
			}
			log.Error("failed to get segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to get segment"))
			return
		}

		log.Info("segment received", slog.String("slug", dbSegment.Slug))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, dbSegment)
	}
}
