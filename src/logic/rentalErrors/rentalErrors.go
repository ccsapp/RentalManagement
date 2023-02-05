// Package rentalErrors defines the semantic rentalErrors which can occur while performing a task process
package rentalErrors

import "errors"

var (
	ErrDomainAssertion         = errors.New("unexpected response from domain service")
	ErrCarNotFound             = errors.New("car not found")
	ErrConflictingRentalExists = errors.New("conflicting rental exists")
)
