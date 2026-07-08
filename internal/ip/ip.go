package ip

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) (string, error) {
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		ip := strings.TrimSpace(cf)
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		ip := strings.TrimSpace(first)
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}
