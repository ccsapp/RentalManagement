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

const (
	invalidTimePeriodMessage = "startDate must be before endDate"
	pastTimePeriodMessage    = "startDate must be in the future"
	carNotFoundMessage       = "car not found"
)

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
	car, err := c.operations.GetCar(ctx.Request().Context(), vin)
	if errors.Is(err, rentalErrors.ErrCarNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, carNotFoundMessage)
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, car)
}

func (c controller) GetNextRental(echo.Context, model.VinParam) error {
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
		return echo.NewHTTPError(http.StatusNotFound, carNotFoundMessage)
	}
	if errors.Is(err, rentalErrors.ErrConflictingRentalExists) {
		return echo.NewHTTPError(http.StatusConflict, "conflicting rental exists")
	}
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusCreated)
}

func (c controller) GetLockState(echo.Context, model.VinParam, model.GetLockStateParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) SetLockState(echo.Context, model.VinParam, model.SetLockStateParams) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GetOverview(ctx echo.Context, params model.GetOverviewParams) error {
	rentals, err := c.operations.GetOverview(ctx.Request().Context(), params.CustomerId)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, *rentals)
}

func (c controller) GetRentalStatus(echo.Context, model.RentalIdParam) error {
	// TODO implement me
	panic("implement me")
}

func (c controller) GrantTrunkAccess(echo.Context, model.RentalIdParam) error {
	// TODO implement me
	panic("implement me")
}

func isInvalidTimePeriod(timePeriod model.TimePeriod) bool {
	return timePeriod.EndDate.Before(timePeriod.StartDate)
}
