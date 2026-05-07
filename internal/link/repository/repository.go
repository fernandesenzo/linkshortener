package repository

import "errors"

var (
	// ideally we would define this error on a place where all repos can share a NotFound error
	// but since the scope of this shortener is very short(lol), lets keep it this way
	ErrNotFound = errors.New("repository: resource not found")
	ErrConflict = errors.New("repository: resource already exists")
)
