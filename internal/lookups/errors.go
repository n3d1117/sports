package lookups

import (
	"errors"
	"fmt"

	"sports/internal/provider/sofascore"
)

type ErrorKind string

const (
	ErrorKindInvalid  ErrorKind = "invalid"
	ErrorKindUpstream ErrorKind = "upstream"
	ErrorKindNotFound ErrorKind = "not_found"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func invalidf(format string, args ...any) error {
	return &Error{
		Kind:    ErrorKindInvalid,
		Message: fmt.Sprintf(format, args...),
	}
}

func translateServiceError(err error) error {
	if err == nil {
		return nil
	}

	var lookupErr *Error
	if errors.As(err, &lookupErr) {
		return err
	}

	var unsupported *sofascoreapi.UnsupportedTournamentEventsError
	if errors.As(err, &unsupported) {
		return &Error{
			Kind:    ErrorKindNotFound,
			Message: fmt.Sprintf("%s events are not available for tournament %d season %d", unsupported.Direction, unsupported.TournamentID, unsupported.SeasonID),
			Err:     err,
		}
	}

	var statusErr *sofascoreapi.HTTPStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case 403:
			return &Error{
				Kind:    ErrorKindUpstream,
				Message: "SofaScore blocked the request with HTTP 403",
				Err:     err,
			}
		case 404:
			return &Error{
				Kind:    ErrorKindNotFound,
				Message: "resource not found",
				Err:     err,
			}
		default:
			return &Error{
				Kind:    ErrorKindUpstream,
				Message: statusErr.Error(),
				Err:     err,
			}
		}
	}

	return &Error{
		Kind:    ErrorKindUpstream,
		Message: err.Error(),
		Err:     err,
	}
}

func eventSectionError(eventID int, section string, err error) error {
	var statusErr *sofascoreapi.HTTPStatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == 404 {
		return &Error{
			Kind:    ErrorKindNotFound,
			Message: fmt.Sprintf("section %q not found for event %d", section, eventID),
			Err:     err,
		}
	}
	return translateServiceError(err)
}

func tournamentSectionError(tournamentID, seasonID int, section string, err error) error {
	var statusErr *sofascoreapi.HTTPStatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == 404 {
		return &Error{
			Kind:    ErrorKindNotFound,
			Message: fmt.Sprintf("section %q not found for tournament %d season %d", section, tournamentID, seasonID),
			Err:     err,
		}
	}
	return translateServiceError(err)
}
