package gethistory

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UserSegmentsHistoryGetter interface {
	GetUserSegmentsHistory(ctx context.Context, id int64, period time.Time) ([][]string, error)
}

// @Summary	Downloading user segments history
// @Tags		users
// @Produce	text/csv,json
// @Param		id		path	string	true	"User ID"
// @Param		period	query	string	true	"Year and month"	example(2023-09)
// @Success	200
// @Failure	400	{object}	resp.ErrResponse
// @Failure	404	{object}	resp.ErrResponse
// @Failure	500	{object}	resp.ErrResponse
// @Router		/users/{id}/download-segments-history [get]
func New(ctx context.Context, log *slog.Logger, userSegmentsHistoryGetter UserSegmentsHistoryGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "httpserver.handlers.users.gethistory.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Info("user id is invalid", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("user id is invalid"))
			return
		}

		period, err := time.Parse("2006-01", r.URL.Query().Get("period"))
		if err != nil {
			log.Info("invalid query param 'period'", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("Invalid query param 'period'. Should be formatted like 'yyyy-mm'"))
			return
		}

		report, err := userSegmentsHistoryGetter.GetUserSegmentsHistory(ctx, id, period)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found")

				render.Render(w, r, resp.ErrNotFound("user not found"))
				return
			}
			log.Error("failed to get user segments history", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to get user segments history"))
			return
		}

		buf := new(bytes.Buffer)
		wtr := csv.NewWriter(buf)
		wtr.WriteAll(report)
		if err := wtr.Error(); err != nil {
			log.Error("failed to write csv", sl.Err(err))

			render.Render(w, r, resp.ErrInternal("failed to write csv"))
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
		w.Header().Set("Content-Type", "text/csv")
		w.Write(buf.Bytes())
	}
}
