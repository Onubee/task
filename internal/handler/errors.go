package handler

import "errors"

var (
	ErrInvalidMethod  = errors.New("invalid HTTP method")
	ErrInternalServer = errors.New("internal server error")
)

type HTTPError struct {
	Status  int         `json:"-"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e HTTPError) Error() string {
	return e.Message
}
