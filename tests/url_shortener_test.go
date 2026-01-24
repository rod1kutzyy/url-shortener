package tests

import (
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/save"
	"github.com/rod1kutzyy/url-shortener/internal/lib/api"
	"github.com/rod1kutzyy/url-shortener/internal/lib/random"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8080"
)

func TestURLShortener_HappyPath(t *testing.T) {
	url := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, url.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().Status(http.StatusCreated).
		JSON().
		Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirectRemove(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		alias    string
		errorMsg string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word(),
		},
		{
			name:     "Invalid URL",
			url:      "invalid_url",
			alias:    gofakeit.Word(),
			errorMsg: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, url.String())

			req := e.POST("/url").
				WithJSON(save.Request{
					URL:   tt.url,
					Alias: tt.alias,
				}).
				WithBasicAuth("myuser", "mypass")

			if tt.errorMsg != "" {
				req.Expect().Status(http.StatusBadRequest).
					JSON().Object().
					Value("error").IsEqual(tt.errorMsg)
				return
			}

			resp := req.Expect().
				Status(http.StatusCreated).
				JSON().Object()

			resp.ContainsKey("alias")

			var alias string
			if tt.alias != "" {
				resp.Value("alias").String().IsEqual(tt.alias)
				alias = tt.alias
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			testRedirect(t, alias, tt.url)

			e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("myuser", "mypass").
				Expect().Status(http.StatusNoContent).
				NoContent()

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	url := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(url.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	url := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(url.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
