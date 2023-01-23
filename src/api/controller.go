package api

import (
	"RentalManagement/logic/model"
	"RentalManagement/logic/operations"
	"github.com/labstack/echo/v4"
)

type controller struct {
	operations operations.IOperations
}

func NewController(operations operations.IOperations) ServerInterface {
	return controller{
		operations,
	}
}

func (c controller) GetAvailableCars(ctx echo.Context, params model.GetAvailableCarsParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetCar(ctx echo.Context, vin model.VinParam) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetNextRental(ctx echo.Context, vin model.VinParam) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) CreateRental(ctx echo.Context, vin model.VinParam, params model.CreateRentalParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetLockState(ctx echo.Context, vin model.VinParam, params model.GetLockStateParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) SetLockState(ctx echo.Context, vin model.VinParam, params model.SetLockStateParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetOverview(ctx echo.Context, params model.GetOverviewParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetRentalStatus(ctx echo.Context, rentalId model.RentalIdParam) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GrantTrunkAccess(ctx echo.Context, rentalId model.RentalIdParam) error {
	// TODO implement me
	panic("implement me")
}
