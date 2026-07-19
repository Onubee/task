package service

import "errors"

var (
	ErrDownloadInProgress = errors.New("download already in progress")
	ErrDownloadTimeout    = errors.New("download timeout exceeded")
	ErrSourceUnavailable  = errors.New("source unavailable")
	ErrEmptyPage          = errors.New("empty page received")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

type DownloadError struct {
	Source  string
	Page    int
	Attempt int
	Err     error
}

func (e DownloadError) Error() string {
	return "source=" + e.Source + ", page=" + string(rune(e.Page)) + ": " + e.Err.Error()
}

func (e DownloadError) Unwrap() error {
	return e.Err
}

func NewDownloadError(source string, page int, attempt int, err error) error {
	return DownloadError{
		Source:  source,
		Page:    page,
		Attempt: attempt,
		Err:     err,
	}
}

func IsRetryable(err error) bool {
	switch {
	case errors.Is(err, ErrSourceUnavailable):
		return true
	case errors.Is(err, ErrDownloadTimeout):
		return true
	case errors.Is(err, ErrRateLimitExceeded):
		return true
	default:
		return false
	}
}
