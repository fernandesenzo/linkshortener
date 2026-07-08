package ip

import (
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name        string
		remoteAddr  string
		cfIP        string
		xff         string
		want        string
		expectError bool
	}{
		{
			name:       "CF-Connecting-IP takes priority",
			remoteAddr: "10.0.0.1:1234",
			cfIP:       "1.2.3.4",
			xff:        "5.6.7.8",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For used when CF header missing",
			remoteAddr: "10.0.0.1:1234",
			xff:        "5.6.7.8, 10.0.0.2",
			want:       "5.6.7.8",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.1:1234",
			want:       "192.168.1.1",
		},
		{
			name:       "CF-Connecting-IP with invalid value falls through to XFF",
			remoteAddr: "10.0.0.1:1234",
			cfIP:       "not-an-ip",
			xff:        "5.6.7.8",
			want:       "5.6.7.8",
		},
		{
			name:       "both headers invalid falls through to RemoteAddr",
			remoteAddr: "192.168.1.1:1234",
			cfIP:       "not-an-ip",
			xff:        "also-not-ip",
			want:       "192.168.1.1",
		},
		{
			name:        "invalid RemoteAddr returns error",
			remoteAddr:  "invalid-addr",
			expectError: true,
		},
		{
			name:   "IPv6 in CF-Connecting-IP",
			cfIP:   "::1",
			remoteAddr: "10.0.0.1:1234",
			want:   "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.cfIP != "" {
				req.Header.Set("CF-Connecting-IP", tt.cfIP)
			}
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			got, err := ClientIP(req)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil and ip=%q", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
