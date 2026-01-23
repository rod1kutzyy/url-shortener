package remove_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/remove"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/remove/mocks"
	"github.com/rod1kutzyy/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/rod1kutzyy/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveHandler(t *testing.T) {
	tests := []struct {
		name       string
		alias      string
		respError  string
		mockError  error
		wantStatus int
	}{
		{
			name:       "Success",
			alias:      "test_alias",
			respError:  "",
			mockError:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Not Found",
			alias:      "non_existent_alias",
			respError:  "not found",
			mockError:  storage.ErrURLNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Internal Error",
			alias:      "test_alias_broken",
			respError:  "internal error",
			mockError:  errors.New("unexpected error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			urlRemoverMock := mocks.NewURLRemover(t)

			urlRemoverMock.On("RemoveURL", tt.alias).
				Return(tt.mockError).Once()

			r := chi.NewRouter()
			r.Delete("/{alias}", remove.New(slogdiscard.NewDiscardLogger(), urlRemoverMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tt.alias, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			respBody := string(body)

			if tt.respError != "" {
				assert.Contains(t, respBody, tt.respError)
			} else {
				assert.Contains(t, respBody, "OK")
			}
		})
	}
}
