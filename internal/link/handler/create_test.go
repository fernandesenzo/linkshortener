package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

type MockService struct {
	CreateLinkFunc func(ctx context.Context, ip string, url string) (*link.Link, error)
	GetLinkFunc    func(ctx context.Context, code string) (*link.Link, error)
}

func (m *MockService) CreateLink(ctx context.Context, ip string, url string) (*link.Link, error) {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(ctx, ip, url)
	}
	return nil, nil
}

func (m *MockService) GetLink(ctx context.Context, code string) (*link.Link, error) {
	if m.GetLinkFunc != nil {
		return m.GetLinkFunc(ctx, code)
	}
	return nil, nil
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		reqBody        string
		remoteAddr     string
		mockSvc        *MockService
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "unsupported content type",
			contentType:    "text/plain",
			reqBody:        `{"url": "https://example.com"}`,
			remoteAddr:     "127.0.0.1:1234",
			mockSvc:        &MockService{},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "unsupported content type\n",
		},
		{
			name:           "invalid request json format",
			contentType:    "application/json",
			reqBody:        `{"url": "https://example.com"`, // broken JSON
			remoteAddr:     "127.0.0.1:1234",
			mockSvc:        &MockService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request\n",
		},
		{
			name:           "invalid request unknown fields",
			contentType:    "application/json",
			reqBody:        `{"url": "https://example.com", "unknown": "field"}`,
			remoteAddr:     "127.0.0.1:1234",
			mockSvc:        &MockService{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request\n",
		},
		{
			name:           "invalid remote addr format",
			contentType:    "application/json",
			reqBody:        `{"url": "https://example.com"}`,
			remoteAddr:     "invalid-addr", // fails SplitHostPort
			mockSvc:        &MockService{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
		{
			name:        "successful link creation",
			contentType: "application/json",
			reqBody:     `{"url": "https://example.com"}`,
			remoteAddr:  "127.0.0.1:1234",
			mockSvc: &MockService{
				CreateLinkFunc: func(ctx context.Context, ip string, url string) (*link.Link, error) {
					if ip != "127.0.0.1" {
						t.Errorf("expected ip '127.0.0.1', got %s", ip)
					}
					if url != "https://example.com" {
						t.Errorf("expected url 'https://example.com', got %s", url)
					}
					return &link.Link{
						Code: "ABCDEF",
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"code":"ABCDEF"}` + "\n",
		},
		{
			name:        "service error ErrInvalidURL",
			contentType: "application/json",
			reqBody:     `{"url": "invalid-url"}`,
			remoteAddr:  "127.0.0.1:1234",
			mockSvc: &MockService{
				CreateLinkFunc: func(ctx context.Context, ip string, url string) (*link.Link, error) {
					return nil, link.ErrInvalidURL
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid url\n",
		},
		{
			name:        "service error ErrTooLongURL",
			contentType: "application/json",
			reqBody:     `{"url": "http://too-long-url..."}`,
			remoteAddr:  "127.0.0.1:1234",
			mockSvc: &MockService{
				CreateLinkFunc: func(ctx context.Context, ip string, url string) (*link.Link, error) {
					return nil, link.ErrTooLongURL
				},
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "url too long\n",
		},
		{
			name:        "service error ErrTooManyActiveURLs",
			contentType: "application/json",
			reqBody:     `{"url": "https://example.com"}`,
			remoteAddr:  "127.0.0.1:1234",
			mockSvc: &MockService{
				CreateLinkFunc: func(ctx context.Context, ip string, url string) (*link.Link, error) {
					return nil, link.ErrTooManyActiveURLs
				},
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "ip already has the limit of links shortened, try again later.\n",
		},
		{
			name:        "service generic error",
			contentType: "application/json",
			reqBody:     `{"url": "https://example.com"}`,
			remoteAddr:  "127.0.0.1:1234",
			mockSvc: &MockService{
				CreateLinkFunc: func(ctx context.Context, ip string, url string) (*link.Link, error) {
					return nil, errors.New("something went wrong")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBufferString(tt.reqBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.RemoteAddr = tt.remoteAddr

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}

			if tt.expectedStatus == http.StatusCreated {
				contentType := rr.Header().Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("expected Content-Type to contain application/json, got %q", contentType)
				}
			}
		})
	}
}
