package link

import (
	"net/url"
	"time"
)

type Link struct {
	OriginalURL string
	Code        string
	ExpiresAt   time.Time
}

const CreateLinkMaxAttempts = 5
const CodeLength = 6
const DefaultTTL = time.Hour * 24
const maxURLlength = 200
const maxActiveLinksForIP = 10

func CanCreate(rawURL string, ipCount int) error {
	if len(rawURL) > maxURLlength {
		return ErrTooLongURL
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrInvalidURL
	}
	if ipCount >= maxActiveLinksForIP {
		return ErrTooManyActiveURLs
	}
	return nil
}
