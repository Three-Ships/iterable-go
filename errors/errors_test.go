package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsForgottenUser(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		expect bool
	}{
		{name: "forgotten user code", err: &ApiError{IterableCode: ITERABLE_ForgottenUserError}, expect: true},
		{name: "wrapped forgotten user code", err: fmt.Errorf("forget: %w", &ApiError{IterableCode: ITERABLE_ForgottenUserError}), expect: true},
		{name: "other code", err: &ApiError{IterableCode: "GenericError"}},
		{name: "non-API error", err: errors.New("ForgottenUserError")},
		{name: "nil", err: nil},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsForgottenUser(tt.err))
		})
	}
}

func TestIsUserNotFound(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		expect bool
	}{
		{name: "lookup code", err: &ApiError{IterableCode: ITERABLE_NoUserWithIdExists}, expect: true},
		{name: "write endpoint message", err: &ApiError{IterableCode: "BadParams", IterableMsg: "User does not exist"}, expect: true},
		{name: "message with different casing and punctuation", err: &ApiError{IterableMsg: "user does not exist."}, expect: true},
		{name: "wrapped", err: fmt.Errorf("forget: %w", &ApiError{IterableMsg: "User does not exist"}), expect: true},
		{name: "other message", err: &ApiError{IterableCode: "BadParams", IterableMsg: "list does not exist"}},
		{name: "forgotten user", err: &ApiError{IterableCode: ITERABLE_ForgottenUserError}},
		{name: "non-API error", err: errors.New("User does not exist")},
		{name: "nil", err: nil},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsUserNotFound(tt.err))
		})
	}
}
