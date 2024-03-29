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

func (c controller) GetNextRental(ctx echo.Context, vin model.VinParam) error {
	rental, err := c.operations.GetNextRental(ctx.Request().Context(), vin)
	if errors.Is(err, rentalErrors.ErrCarNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, carNotFoundMessage)
	}
	if err != nil {
		return err
	}
	if rental == nil {
		return ctx.NoContent(http.StatusNoContent)
	}
	return ctx.JSON(http.StatusOK, rental)
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

func (c controller) GetLockState(ctx echo.Context, vin model.VinParam, params model.GetLockStateParams) error {
	lockState, err := c.operations.GetLockState(ctx.Request().Context(), vin, params.TrunkAccessToken)

	if errors.Is(err, rentalErrors.ErrTrunkAccessDenied) {
		return echo.NewHTTPError(http.StatusForbidden, "trunk access denied")
	}

	if err != nil {
		return err
	}

	lockStateObject := model.LockStateObject{
		TrunkLockState: *lockState,
	}

	return ctx.JSON(http.StatusOK, lockStateObject)
}

func (c controller) SetLockState(ctx echo.Context, vin model.VinParam, params model.SetLockStateParams) error {
	if params.CustomerId == nil && params.TrunkAccessToken == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "either customerId or trunkAccessToken must be specified")
	}

	if params.CustomerId != nil && params.TrunkAccessToken != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "only one of customerId or trunkAccessToken can be specified")
	}

	var lockStateObject model.LockStateObject

	// bind errors are unexpected because the trunkLockState is validated by the Swagger spec
	err := ctx.Bind(&lockStateObject)

	if err != nil {
		return err
	}

	lockState := lockStateObject.TrunkLockState

	if params.CustomerId != nil {
		err = c.operations.SetLockStateCustomerId(ctx.Request().Context(), lockState, vin, *params.CustomerId)
	} else if params.TrunkAccessToken != nil {
		err = c.operations.SetLockStateTrunkAccessToken(ctx.Request().Context(), lockState,
			vin, *params.TrunkAccessToken)
	}

	if errors.Is(err, rentalErrors.ErrTrunkAccessDenied) {
		return echo.NewHTTPError(http.StatusForbidden, "trunk access denied")
	}

	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (c controller) GetOverview(ctx echo.Context, params model.GetOverviewParams) error {
	rentals, err := c.operations.GetOverview(ctx.Request().Context(), params.CustomerId)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, *rentals)
}

func (c controller) GetRentalStatus(ctx echo.Context, rentalId model.RentalIdParam) error {
	rental, err := c.operations.GetRentalStatus(ctx.Request().Context(), rentalId)
	if errors.Is(err, rentalErrors.ErrRentalNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "rentalId not found")
	}
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, *rental)
}

func (c controller) GrantTrunkAccess(ctx echo.Context, rentalId model.RentalIdParam) error {
	var timePeriod model.TimePeriod
	// bind errors are unexpected because the timePeriod is validated by the Swagger spec
	err := ctx.Bind(&timePeriod)
	if err != nil {
		return err
	}

	if isInvalidTimePeriod(timePeriod) {
		return echo.NewHTTPError(http.StatusBadRequest, invalidTimePeriodMessage)
	}

	trunkAccess, err := c.operations.GrantTrunkAccess(ctx.Request().Context(), rentalId, timePeriod)
	if errors.Is(err, rentalErrors.ErrRentalNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "rental not found")
	}
	if errors.Is(err, rentalErrors.ErrRentalNotActive) {
		return echo.NewHTTPError(http.StatusForbidden, "rental not active")
	}
	if errors.Is(err, rentalErrors.ErrRentalNotOverlapping) {
		return echo.NewHTTPError(http.StatusForbidden, "rental not overlapping")
	}
	if errors.Is(err, rentalErrors.ErrResourceConflict) {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "failed to grant trunk access")
	}
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, trunkAccess)
}

func isInvalidTimePeriod(timePeriod model.TimePeriod) bool {
	return timePeriod.EndDate.Before(timePeriod.StartDate)
}
