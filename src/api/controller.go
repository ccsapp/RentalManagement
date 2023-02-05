package api

import (
	"RentalManagement/logic/model"
	"RentalManagement/logic/operations"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/util"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
)

const invalidTimePeriodMessage = "startDate must be before endDate"
const pastTimePeriodMessage = "startDate must be in the future"

type controller struct {
	operations   operations.IOperations
	timeProvider util.ITimeProvider
}

func NewController(operations operations.IOperations, timeProvider util.ITimeProvider) ServerInterface {
	return controller{
		operations,
		timeProvider,
	}
}

func (c controller) GetAvailableCars(ctx echo.Context, params model.GetAvailableCarsParams) error {
	if isInvalidTimePeriod(params.TimePeriod) {
		return echo.NewHTTPError(http.StatusBadRequest, invalidTimePeriodMessage)
	}
	cars, err := c.operations.GetAvailableCars(ctx.Request().Context(), params.TimePeriod)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, cars)
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
	var timePeriod model.TimePeriod
	// bind errors are unexpected because the timePeriod is validated by the Swagger spec
	err := ctx.Bind(&timePeriod)
	if err != nil {
		return err
	}

	if isInvalidTimePeriod(timePeriod) {
		return echo.NewHTTPError(http.StatusBadRequest, invalidTimePeriodMessage)
	}
	if timePeriod.StartDate.Before(c.timeProvider.Now()) {
		return echo.NewHTTPError(http.StatusForbidden, pastTimePeriodMessage)
	}

	err = c.operations.CreateRental(ctx.Request().Context(), vin, params.CustomerId, timePeriod)
	if errors.Is(err, rentalErrors.ErrCarNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "car not found")
	}
	if errors.Is(err, rentalErrors.ErrConflictingRentalExists) {
		return echo.NewHTTPError(http.StatusConflict, "conflicting rental exists")
	}
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusCreated)
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

func isInvalidTimePeriod(timePeriod model.TimePeriod) bool {
	return timePeriod.EndDate.Before(timePeriod.StartDate)
}
