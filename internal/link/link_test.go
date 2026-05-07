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
			name:    "valid request",
			url:     "https://google.com",
			ipCount: 5,
			wantErr: nil,
		},
		{
			name:    "url too long",
			url:     strings.Repeat("a", 201),
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
			name:    "ip limit exceeded",
			url:     "https://google.com",
			ipCount: 11,
			wantErr: link.ErrTooManyActiveURLs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := link.CanCreate(tt.url, tt.ipCount)
			if err != tt.wantErr {
				t.Errorf("CanCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
