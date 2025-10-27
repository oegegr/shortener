// Package errors содержит ошибки, связанные с сервисом сокращения URL-адресов.
package errors

import (
	"errors"
)

// ErrServiceFailedToGetShortURL представляет ошибку, которая возникает при неудаче в получении сокращенного URL-адреса.
var ErrServiceFailedToGetShortURL = errors.New("failed to get short url")

// ErrServiceURLGone представляет ошибку, которая возникает при попытке доступа к удаленному URL-адресу.
var ErrServiceURLGone = errors.New("url has been deleted")
