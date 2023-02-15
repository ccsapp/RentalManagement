package operations

import (
	"RentalManagement/logic/model"
	"context"
)

//go:generate mockgen -source=interface.go -package=mocks -destination=../../mocks/mock_operations.go

type IOperations interface {
	//GetAvailableCars Get Available Cars in a Time Period
	GetAvailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.CarAvailable, error)
	// CreateRental Create a New Rental
	CreateRental(ctx context.Context, vin model.Vin, customerID model.CustomerId, timePeriod model.TimePeriod) error
	GetCar(ctx context.Context, vin model.Vin) (*model.Car, error)
	// GetOverview Get an Overview of a Customer’s Rentals
	GetOverview(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error)
	// GetRentalStatus Get Rental Status Information (Including Car Data) based on an ID
	GetRentalStatus(ctx context.Context, rentalId model.RentalId) (*model.Rental, error)
}
