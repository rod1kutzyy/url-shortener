package redirect_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/redirect"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/redirect/mocks"
	"github.com/rod1kutzyy/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/rod1kutzyy/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:      "Success",
			alias:     "test_alias",
			url:       "https://www.google.com",
			respError: "",
			mockError: nil,
		},
		{
			name:      "Not Found",
			alias:     "test_alias_not_found",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "Internal Error",
			alias:     "test_alias_internal_error",
			respError: "internal error",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewURLGetter(t)

			urlGetterMock.On("GetURL", tt.alias).
				Return(tt.url, tt.mockError).Once()

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			client := ts.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			resp, err := client.Get(ts.URL + "/" + tt.alias)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			respBody := string(body)

			if tt.respError != "" {
				require.NotEqual(t, http.StatusFound, resp.StatusCode)
				assert.Contains(t, respBody, tt.respError)
			} else {
				require.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, tt.url, resp.Header.Get("Location"))
			}
		})
	}
}
