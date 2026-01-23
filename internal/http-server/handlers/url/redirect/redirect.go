package redirect

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

//go:generate mockery --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(logger *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			logger.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				logger.Info("url not found", "alias", alias)
				render.JSON(w, r, response.Error("not found"))
				return
			}
			logger.Error("failed to get url", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		logger.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
