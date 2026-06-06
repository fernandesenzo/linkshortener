package link

import "errors"

var (
	ErrTooLongURL        = errors.New("the provided URL is too long. the maximum allowed size is 200")
	ErrTooManyActiveURLs = errors.New("too many active URLs created from your IP.")
	ErrTooManyCollisions = errors.New("could not create link because it reached maximum collisions")
	ErrInvalidURL        = errors.New("invalid url")
	ErrNotFound          = errors.New("no link found with such code")
)
