package remove

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rod1kutzyy/url-shortener/internal/lib/api/response"
	"github.com/rod1kutzyy/url-shortener/internal/lib/logger/sl"
	"github.com/rod1kutzyy/url-shortener/internal/storage"
)

//go:generate mockery --name=URLRemover
type URLRemover interface {
	RemoveURL(alias string) error
}

func New(logger *slog.Logger, urlRemover URLRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.remove.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			logger.Info("alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		err := urlRemover.RemoveURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				logger.Info("url not found", "alias", alias)
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.Error("not found"))
				return
			}
			logger.Error("failed to remove url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		logger.Info("url removed", slog.String("alias", alias))

		render.NoContent(w, r)
	}
}
