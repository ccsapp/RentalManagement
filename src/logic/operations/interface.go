package operations

import (
	"RentalManagement/logic/model"
	"context"
)

//go:generate mockgen -source=interface.go -package=mocks -destination=../../mocks/mock_operations.go

type IOperations interface {
	GetAvailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.CarAvailable, error)
	CreateRental(ctx context.Context, vin model.Vin, customerID model.CustomerId, timePeriod model.TimePeriod) error
	GetCar(ctx context.Context, vin model.Vin) (*model.Car, error)
}
