package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SegmentDeleter interface {
	DeleteSegment(ctx context.Context, slug string) error
}

// @Summary	Deleting a segment
// @Tags		segments
// @Param		slug	path	string	true	"Segment slug"
// @Success	204
// @Failure	400	{object}	resp.ErrResponse
// @Failure	404	{object}	resp.ErrResponse
// @Failure	500	{object}	resp.ErrResponse
// @Router		/segments/{slug} [delete]
func New(ctx context.Context, log *slog.Logger, segmentDeleter SegmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.delete.New"

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

		err := segmentDeleter.DeleteSegment(ctx, slug)
		if err != nil {
			if errors.Is(err, storage.ErrSegmentNotFound) {
				log.Info("segment not found", slog.String("slug", slug))

				render.Render(w, r, resp.ErrNotFound("segment not found"))
				return
			}
			log.Error("failed to delete segment", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to delete segment"))
			return
		}

		log.Info("segment deleted", slog.String("slug", slug))

		w.WriteHeader(http.StatusNoContent)
	}
}
