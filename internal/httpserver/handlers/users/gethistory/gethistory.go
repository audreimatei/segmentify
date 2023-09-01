package gethistory

import (
	"bytes"
	"encoding/csv"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"segmentify/internal/lib/logger/sl"
	resp "segmentify/internal/lib/response"
	"segmentify/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type UserSegmentsHistoryGetter interface {
	GetUserSegmentsHistory(userID int64, period time.Time) ([][]string, error)
}

func New(log *slog.Logger, userSegmentsHistoryGetter UserSegmentsHistoryGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "httpserver.handlers.users.gethistory.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			log.Info("userID is invalid", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("user_id is invalid"))
			return
		}

		period, err := time.Parse("2006-01", r.URL.Query().Get("period"))
		if err != nil {
			log.Info("invalid query param 'period'", sl.Err(err))

			render.Render(w, r, resp.ErrInvalidRequest("Invalid query param 'period'. Should be formatted like 'yyyy-mm'"))
			return
		}

		report, err := userSegmentsHistoryGetter.GetUserSegmentsHistory(userID, period)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found")

				render.Render(w, r, resp.ErrInvalidRequest("user not found"))
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