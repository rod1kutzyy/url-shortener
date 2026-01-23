package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/save"
	"github.com/rod1kutzyy/url-shortener/internal/http-server/handlers/url/save/mocks"
	"github.com/rod1kutzyy/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/rod1kutzyy/url-shortener/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	tests := []struct {
		name       string
		alias      string
		url        string
		respError  string
		mockError  error
		wantStatus int
	}{
		{
			name:       "Success",
			alias:      "test_alias",
			url:        "https://google.com",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Empty alias",
			alias:      "",
			url:        "https://google.com",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Empty url",
			alias:      "some_alias",
			url:        "",
			respError:  "field URL is required",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid URL",
			alias:      "test_alias",
			url:        "some invalid URL",
			respError:  "field URL is not a valid URL",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "SaveURL Error",
			alias:      "test_alias",
			url:        "https://google.com",
			respError:  "failed to add url",
			mockError:  errors.New("unexpected error"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Alias Exists",
			alias:      "test_alias",
			url:        "https://google.com",
			respError:  "url already exists",
			mockError:  storage.ErrURLExists,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tt.respError == "" || tt.mockError != nil {
				urlSaverMock.On("SaveURL", tt.url, mock.AnythingOfType("string")).
					Return(int64(1), tt.mockError).Once()
			}

			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tt.url, tt.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, resp.Error, tt.respError)
		})
	}
}
