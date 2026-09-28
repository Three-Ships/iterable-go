package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	STAGE_BEFORE_REQUEST = "before-request"
	STAGE_REQUEST        = "request"
	STAGE_AFTER_REQUEST  = "after-request"

	TYPE_UNKNOWN         = "unknown"
	TYPE_NOT_IMPLEMENTED = "not-implemented"
	TYPE_JSON_PARSE      = "json"
	TYPE_REQUEST_PREP    = "request-prep"
	TYPE_IO              = "io"
	TYPE_HTTP_STATUS     = "not-ok-http-status"
	TYPE_INVALID_DATA    = "invalid-data"

	ITERABLE_NoUserWithIdExists      = "error.users.noUserWithIdExists"
	ITERABLE_InvalidList             = "error.lists.invalidListId"
	ITERABLE_Success                 = "Success"
	ITERABLE_FieldTypeMismatchErrStr = "RequestFieldsTypesMismatched"
	ITERABLE_ForgottenUserError      = "ForgottenUserError"
)

// userDoesNotExistMsg is how Iterable reports an unknown user on write
// endpoints such as users/forget. It arrives in the "msg" field under a
// generic code, so it can only be matched on the message text.
const userDoesNotExistMsg = "user does not exist"

type ApiError struct {
	Stage          string
	Type           string
	SourceErr      error
	Body           []byte
	HttpStatusCode int

	IterableCode string
	// IterableMsg is the "msg" field of an Iterable error response.
	IterableMsg string
}

var _ error = &ApiError{}

func (e *ApiError) Error() string {
	var err string
	if e.SourceErr != nil {
		err = e.SourceErr.Error()
	} else {
		err = string(e.Body)
	}
	return fmt.Sprintf(
		"http request to Iterable failed during '%s' stage with error type '%s', httpStatus: '%d'; original err: %v",
		e.Stage, e.Type, e.HttpStatusCode, err,
	)
}

// IsForgottenUser reports whether Iterable rejected a request because the
// user was forgotten (GDPR). Iterable rejects every write for a forgotten user.
func IsForgottenUser(err error) bool {
	var apiErr *ApiError
	return errors.As(err, &apiErr) && apiErr.IterableCode == ITERABLE_ForgottenUserError
}

// IsUserNotFound reports whether Iterable has no user matching the request.
// Lookup endpoints signal this with ITERABLE_NoUserWithIdExists; write
// endpoints such as users/forget signal it only in the message text.
func IsUserNotFound(err error) bool {
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.IterableCode == ITERABLE_NoUserWithIdExists ||
		strings.Contains(strings.ToLower(apiErr.IterableMsg), userDoesNotExistMsg)
}

// Is method is required by errors.Is() to properly distinguish between
// different types -vs- same pointer to the same type.
// Without it, errors.Is(err, ErrFieldTypeMismatch) returns false:
// ok := errors.Is(errors.Join(&iterable_errors.ApiError{}), &iterable_errors.ApiError{})
// ^ would be false
func (e *ApiError) Is(other error) bool {
	var err *ApiError
	return errors.As(other, &err) && err != nil
}
