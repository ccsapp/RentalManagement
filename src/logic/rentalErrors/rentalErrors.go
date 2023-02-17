// Package rentalErrors defines the semantic rentalErrors which can occur while performing a task process
package rentalErrors

import "errors"

var (
	ErrDomainAssertion         = errors.New("unexpected response from domain service")
	ErrCarNotFound             = errors.New("car not found")
	ErrConflictingRentalExists = errors.New("conflicting rental exists")
	ErrRentalNotFound          = errors.New("rental not found")
	ErrRentalNotActive         = errors.New("rental not active")
	ErrRentalNotOverlapping    = errors.New("rental does not overlap the requested time period")
	// ErrResourceConflict is returned when a resource is already in use and retry attempts failed.
	ErrResourceConflict = errors.New("resource conflict")
)
