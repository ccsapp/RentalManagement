package operations

import (
	"RentalManagement/logic/model"
	"context"
)

//go:generate mockgen -source=interface.go -package=mocks -destination=../../mocks/mock_operations.go

type IOperations interface {
	// GetAvailableCars Get Available Cars in a Time Period
	GetAvailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.CarAvailable, error)
	// CreateRental Create a New Rental
	CreateRental(ctx context.Context, vin model.Vin, customerID model.CustomerId, timePeriod model.TimePeriod) error
	// GetNextRental Get the active or next upcoming Rental of a Car in a format suitable for the Fleet Manager, that is,
	// the active status, the customer, the rental period, and the rental ID.
	// Returns nil if there is no next rental.
	GetNextRental(ctx context.Context, vin model.Vin) (*model.Rental, error)
	GetCar(ctx context.Context, vin model.Vin) (*model.Car, error)
	// GetOverview Get an Overview of a Customer’s Rentals
	GetOverview(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error)
	// GetRentalStatus Get Rental Status Information (Including Car Data) based on an ID
	GetRentalStatus(ctx context.Context, rentalId model.RentalId) (*model.Rental, error)
	// GrantTrunkAccess Generate a new Trunk Access Token and replace the old one of the rental
	// with given rentalId with it, if present. The new access token is returned.
	// Returns rentalErrors.ErrRentalNotFound if the rental does not exist.
	// Returns rentalErrors.ErrRentalNotActive if the rental is not active.
	// Returns rentalErrors.ErrRentalNotOverlapping if the rental is not active at any time during the validity period.
	// Returns rentalErrors.ErrResourceConflict if the resource is already in use and retry attempts failed.
	GrantTrunkAccess(ctx context.Context, rentalId model.RentalId, timePeriod model.TimePeriod) (
		*model.TrunkAccess, error)
	// GetLockState Get TrunkLockState of a Car if valid token is provided
	// Returns rentalErrors.ErrTrunkAccessDenied if the token is not valid
	// Returns rentalErrors.ErrDomainAssertion if communication with the domain microservice
	// did not return the current trunk lock state
	GetLockState(ctx context.Context, vin model.Vin, token model.TrunkAccessToken) (*model.LockState, error)
	// SetLockStateCustomerId Set TrunkLockState of a Car if customer has an active rental for the car
	// Returns rentalErrors.ErrTrunkAccessDenied if the customer does not have an active rental for the car
	SetLockStateCustomerId(ctx context.Context, lockState model.LockState, vin model.Vin,
		customerId model.CustomerId) error
	// SetLockStateTrunkAccessToken Set TrunkLockState of a Car if a valid token is provided
	// Returns rentalErrors.ErrTrunkAccessDenied if the token is not valid
	SetLockStateTrunkAccessToken(ctx context.Context, lockState model.LockState, vin model.Vin,
		token model.TrunkAccessToken) error
}
