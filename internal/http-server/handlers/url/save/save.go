package save

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/rod1kutzyy/url-shortener/internal/lib/api/response"
	"github.com/rod1kutzyy/url-shortener/internal/lib/logger/sl"
	"github.com/rod1kutzyy/url-shortener/internal/lib/random"
	"github.com/rod1kutzyy/url-shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

//go:generate mockery --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(logger *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			logger.Error("invalid request", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				logger.Info("url already exists", slog.String("url", req.URL))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, response.Error("url already exists"))
				return
			}
			logger.Error("failed to add url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}

		logger.Info("url added", slog.Int64("id", id))

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
