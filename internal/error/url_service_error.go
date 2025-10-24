package errors

import (
	"errors"
)

var (
	ErrServiceFailedToGetShortURL = errors.New("failed to get short url")
	ErrServiceURLGone             = errors.New("url has been deleted")
)
