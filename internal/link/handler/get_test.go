package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

func TestHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		mockSvc        *MockService
		expectedStatus int
		expectedBody   string
		expectedLoc    string // for redirect header
	}{
		{
			name:           "invalid code length",
			code:           "ABC", // length 3 (not 6)
			mockSvc:        &MockService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request\n",
		},
		{
			name: "successful redirect",
			code: "ABCDEF",
			mockSvc: &MockService{
				GetLinkFunc: func(ctx context.Context, code string) (*link.Link, error) {
					if code != "ABCDEF" {
						t.Errorf("expected code 'ABCDEF', got %s", code)
					}
					return &link.Link{
						Code:        "ABCDEF",
						OriginalURL: "https://example.com/dest",
					}, nil
				},
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedLoc:    "https://example.com/dest",
		},
		{
			name: "link not found",
			code: "NOTFND",
			mockSvc: &MockService{
				GetLinkFunc: func(ctx context.Context, code string) (*link.Link, error) {
					return nil, link.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "no link with this code\n",
		},
		{
			name: "internal service error",
			code: "DBERR1",
			mockSvc: &MockService{
				GetLinkFunc: func(ctx context.Context, code string) (*link.Link, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)
			req.SetPathValue("code", tt.code)

			rr := httptest.NewRecorder()
			h.Get(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			}

			if tt.expectedLoc != "" {
				loc := rr.Header().Get("Location")
				if loc != tt.expectedLoc {
					t.Errorf("expected Location header %q, got %q", tt.expectedLoc, loc)
				}
			}
		})
	}
}
