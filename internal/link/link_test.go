package link_test

import (
	"strings"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

func TestCanCreate(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		ipCount int
		wantErr error
	}{
		{
			name:    "valid https request",
			url:     "https://google.com",
			ipCount: 5,
			wantErr: nil,
		},
		{
			name:    "valid http request",
			url:     "http://localhost:8080/test",
			ipCount: 0,
			wantErr: nil,
		},
		{
			name:    "url too long",
			url:     "https://" + strings.Repeat("a", 201),
			ipCount: 0,
			wantErr: link.ErrTooLongURL,
		},
		{
			name:    "ip limit reached",
			url:     "https://google.com",
			ipCount: 10,
			wantErr: link.ErrTooManyActiveURLs,
		},
		{
			name:    "missing protocol",
			url:     "google.com",
			ipCount: 0,
			wantErr: link.ErrInvalidURL,
		},
		{
			name:    "invalid protocol (ftp)",
			url:     "ftp://files.com",
			ipCount: 0,
			wantErr: link.ErrInvalidURL,
		},
		{
			name:    "malicious protocol (javascript)",
			url:     "javascript:alert(1)",
			ipCount: 0,
			wantErr: link.ErrInvalidURL,
		},
		{
			name:    "empty string",
			url:     "",
			ipCount: 0,
			wantErr: link.ErrInvalidURL,
		},
		{
			name:    "garbage string",
			url:     "this is not a url",
			ipCount: 0,
			wantErr: link.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := link.CanCreate(tt.url, tt.ipCount)
			if err != tt.wantErr {
				t.Errorf("CanCreate() for %s: error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
