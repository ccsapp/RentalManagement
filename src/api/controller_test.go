package api

import (
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/mocks"
	"RentalManagement/testdata"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	carTypes "github.com/ccsapp/cargotypes"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var exampleCustomerID = "customer@example.com"

var availableCar1 = model.CarAvailable{
	Brand:         "Audi",
	Model:         "A3",
	NumberOfSeats: 7,
	Vin:           "WVWAA71K08W201030",
}

var availableCar2 = model.CarAvailable{
	Brand:         "Mercedes",
	Model:         "B4",
	NumberOfSeats: 9,
	Vin:           "1FVNY5Y90HP312888",
}

var staticCar = model.Car{
	Brand: "Tesla",
	Model: "Model X",
	TechnicalSpecification: &model.TechnicalSpecification{
		Color: "black",
		Consumption: carTypes.TechnicalSpecificationConsumption{
			City:     10,
			Combined: 12,
			Overland: 13,
		},
		Emissions: carTypes.TechnicalSpecificationEmissions{
			City:     17,
			Combined: 18,
			Overland: 19,
		},
		Engine: carTypes.TechnicalSpecificationEngine{
			Power: 25,
			Type:  "180 CDI",
		},
		Fuel:          model.HYBRIDDIESEL,
		FuelCapacity:  "54.0L;85.2kWh",
		NumberOfDoors: 4,
		NumberOfSeats: 5,
		Transmission:  model.MANUAL,
		TrunkVolume:   1000,
		Weight:        2000},
	Vin: "1FVNY5Y90HP312888",
}

var exampleRental = model.Rental{
	State:    model.UPCOMING,
	Customer: &model.Customer{CustomerId: exampleCustomerID},
	Car:      &staticCar,
	Id:       "kskgnvsl",
	RentalPeriod: model.TimePeriod{
		StartDate: time.Date(2123, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2123, 3, 1, 0, 0, 0, 0, time.UTC),
	},
}

var timePeriod = model.TimePeriod{
	StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
}

var invalidTimePeriod = model.TimePeriod{
	StartDate: time.Date(2099, 2, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(1999, 3, 1, 0, 0, 0, 0, time.UTC),
}

var timePeriod1900 = model.TimePeriod{
	StartDate: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(1900, 12, 31, 23, 59, 59, 999999, time.UTC),
}

var availableCars = []model.CarAvailable{availableCar1, availableCar2}

var currentTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
var future = time.Date(2090, 2, 1, 0, 0, 0, 0, time.UTC)

var carBase1 = model.Car{
	Vin:   "3VW217AU9FM500158",
	Brand: "Volkswagen",
	Model: "Golf",
}

var carBase2 = model.Car{
	Vin:   "WVWAA71K08W201030",
	Brand: "Audi",
	Model: "A3",
}

var rentalCustomerShort1 = model.Rental{
	State: model.EXPIRED,
	Car:   &carBase1,
	Id:    "M9hUnd8a",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(1900, 3, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 4, 3, 1, 0, 0, 0, time.UTC),
	},
}

var rentalCustomerShort2 = model.Rental{
	State: model.ACTIVE,
	Car:   &carBase2,
	Id:    "P2zUdL3C",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 1, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
}

var customerRentalsShort = []model.Rental{rentalCustomerShort1, rentalCustomerShort2}

var trunkAccess = model.TrunkAccess{
	Token: "bumrLuCMbumrLuCMbumrLuCM",
	ValidityPeriod: model.TimePeriod{
		StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

func TestController_GetAvailableCars_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)
	mockEchoContext.EXPECT().JSON(http.StatusOK, &availableCars)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetAvailableCars(ctx, timePeriod).Return(&availableCars, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetAvailableCars(mockEchoContext, model.GetAvailableCarsParams{TimePeriod: timePeriod})
	assert.Nil(t, err)
}

func TestController_GetAvailableCars_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	operationsError := errors.New("operations error")

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetAvailableCars(ctx, timePeriod).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetAvailableCars(mockEchoContext, model.GetAvailableCarsParams{TimePeriod: timePeriod})
	assert.ErrorIs(t, err, operationsError)
}

func TestController_GetAvailableCars_InvalidTimePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)
	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetAvailableCars(mockEchoContext, model.GetAvailableCarsParams{TimePeriod: invalidTimePeriod})

	assert.Equal(t, echo.NewHTTPError(http.StatusBadRequest, "startDate must be before endDate"), err)
}

func TestController_GetCar_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)
	mockEchoContext.EXPECT().JSON(http.StatusOK, &staticCar)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetCar(ctx, testdata.VinCar).Return(&staticCar, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetCar(mockEchoContext, testdata.VinCar)
	assert.Nil(t, err)
}

func TestController_GetCar_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().GetCar(ctx, testdata.VinCar).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetCar(mockEchoContext, testdata.VinCar)
	assert.ErrorIs(t, err, operationsError)
}

func TestController_GetCar_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetCar(ctx, testdata.VinCar).Return(nil, rentalErrors.ErrCarNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetCar(mockEchoContext, testdata.VinCar)
	assert.Equal(t, echo.NewHTTPError(http.StatusNotFound, "car not found"), err)
}

func TestController_GetNextRental_success_exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)
	mockEchoContext.EXPECT().JSON(http.StatusOK, &exampleRental)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetNextRental(ctx, testdata.VinCar).Return(&exampleRental, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetNextRental(mockEchoContext, testdata.VinCar)
	assert.Nil(t, err)
}

func TestController_GetNextRental_success_notExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)
	mockEchoContext.EXPECT().NoContent(http.StatusNoContent)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetNextRental(ctx, testdata.VinCar).Return(nil, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetNextRental(mockEchoContext, testdata.VinCar)
	assert.Nil(t, err)
}

func TestController_GetNextRental_UnknownCar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetNextRental(ctx, testdata.VinCar).Return(nil, rentalErrors.ErrCarNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetNextRental(mockEchoContext, testdata.VinCar)
	assert.Equal(t, echo.NewHTTPError(http.StatusNotFound, "car not found"), err)
}

func TestController_GetNextRental_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	operationsError := errors.New("operations error")

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetNextRental(ctx, testdata.VinCar).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetNextRental(mockEchoContext, testdata.VinCar)
	assert.ErrorIs(t, err, operationsError)
}

func TestController_CreateRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)
	mockEchoContext.EXPECT().NoContent(http.StatusCreated)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).Return(nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(currentTime)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})
	assert.Nil(t, err)
}

func TestController_CreateRental_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	operationsError := errors.New("operations error")

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).Return(operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(currentTime)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})
	assert.ErrorIs(t, err, operationsError)
}

func TestController_CreateRental_CarNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).
		Return(rentalErrors.ErrCarNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(currentTime)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})
	assert.Equal(t, echo.NewHTTPError(http.StatusNotFound, "car not found"), err)
}

func TestController_CreateRental_invalidTimePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, invalidTimePeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, echo.NewHTTPError(http.StatusBadRequest, "startDate must be before endDate"), err)
}

func TestController_CreateRental_Past(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod1900).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(future)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "startDate must be in the future"), err)
}

func TestController_CreateRental_ConflictingRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).
		Return(rentalErrors.ErrConflictingRentalExists)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(currentTime)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, echo.NewHTTPError(http.StatusConflict, "conflicting rental exists"), err)
}

func TestController_GetOverview_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().JSON(http.StatusOK, customerRentalsShort)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetOverview(ctx, exampleCustomerID).Return(&customerRentalsShort, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetOverview(mockContext, model.GetOverviewParams{CustomerId: exampleCustomerID})
	assert.Nil(t, err)
}

func TestController_GetOverview_operationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	operationsError := errors.New("operations error")

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetOverview(ctx, exampleCustomerID).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetOverview(mockContext, model.GetOverviewParams{CustomerId: exampleCustomerID})
	assert.ErrorIs(t, err, operationsError)
}

func TestController_GrantTrunkAccess_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)
	mockContext.EXPECT().JSON(http.StatusCreated, &trunkAccess)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(&trunkAccess, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Nil(t, err)
}

func TestController_GrantTrinkAccess_invalidTimePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, invalidTimePeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Equal(t, echo.NewHTTPError(http.StatusBadRequest, "startDate must be before endDate"), err)
}

func TestController_GrantTrunkAccess_rentalNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(nil, rentalErrors.ErrRentalNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Equal(t, echo.NewHTTPError(http.StatusNotFound, "rental not found"), err)
}

func TestController_GrantTrunkAccess_rentalNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(nil, rentalErrors.ErrRentalNotActive)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "rental not active"), err)
}

func TestController_GrantTrunkAccess_rentalNotOverlapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(nil, rentalErrors.ErrRentalNotOverlapping)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "rental not overlapping"), err)
}

func TestController_GrantTrunkAccess_resourceConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(nil, rentalErrors.ErrResourceConflict)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.Equal(t, echo.NewHTTPError(http.StatusServiceUnavailable, "failed to grant trunk access"), err)
}

func TestController_GrantTrunkAccess_operationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, trunkAccess.ValidityPeriod).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().GrantTrunkAccess(ctx, "rentalId", trunkAccess.ValidityPeriod).
		Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GrantTrunkAccess(mockContext, "rentalId")

	assert.ErrorIs(t, err, operationsError)
}

func TestController_GetRentalStatus_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().JSON(http.StatusOK, rentalCustomerShort1)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetRentalStatus(ctx, rentalCustomerShort1.Id).Return(&rentalCustomerShort1, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetRentalStatus(mockContext, rentalCustomerShort1.Id)
	assert.Nil(t, err)
}

func TestController_GetRentalStatus_RentalNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetRentalStatus(ctx, rentalCustomerShort1.Id).Return(nil, rentalErrors.ErrRentalNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetRentalStatus(mockContext, rentalCustomerShort1.Id)

	assert.Equal(t, echo.NewHTTPError(http.StatusNotFound, "rentalId not found"), err)
}

func TestController_GetRentalStatus_operationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().GetRentalStatus(ctx, rentalCustomerShort1.Id).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetRentalStatus(mockContext, rentalCustomerShort1.Id)
	assert.ErrorIs(t, err, operationsError)
}

func TestController_GetLockState_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	locked := model.LOCKED

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().JSON(http.StatusOK, model.LockStateObject{TrunkLockState: locked})

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().GetLockState(ctx, testdata.VinCar, testdata.TrunkAccessToken).Return(&locked, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetLockState(mockContext, testdata.VinCar, model.GetLockStateParams{TrunkAccessToken: testdata.TrunkAccessToken})
	assert.Nil(t, err)
}

func TestController_GetLockState_TrunkAccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)

	mockOperations.EXPECT().GetLockState(ctx, testdata.VinCar, testdata.TrunkAccessToken).Return(nil, rentalErrors.ErrTrunkAccessDenied)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetLockState(mockContext, testdata.VinCar, model.GetLockStateParams{TrunkAccessToken: testdata.TrunkAccessToken})
	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "trunk access denied"), err)
}

func TestController_GetLockState_operationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().GetLockState(ctx, testdata.VinCar, testdata.TrunkAccessToken).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetLockState(mockContext, testdata.VinCar, model.GetLockStateParams{TrunkAccessToken: testdata.TrunkAccessToken})
	assert.ErrorIs(t, err, operationsError)
}

func TestController_SetLockState_customerId_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)
	mockContext.EXPECT().NoContent(http.StatusNoContent)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().SetLockStateCustomerId(ctx, model.LOCKED, testdata.VinCar, exampleCustomerID).Return(nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       &exampleCustomerID,
		TrunkAccessToken: nil,
	})
	assert.Nil(t, err)
}

func TestController_SetLockState_trunkAccessToken_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)
	mockContext.EXPECT().NoContent(http.StatusNoContent)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().SetLockStateTrunkAccessToken(ctx, model.LOCKED, testdata.VinCar, testdata.TrunkAccessToken).Return(nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	token := testdata.TrunkAccessToken
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       nil,
		TrunkAccessToken: &token,
	})
	assert.Nil(t, err)
}

func TestController_SetLockState_customerId_TrunkAccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().SetLockStateCustomerId(ctx, model.LOCKED, testdata.VinCar, exampleCustomerID).Return(rentalErrors.ErrTrunkAccessDenied)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       &exampleCustomerID,
		TrunkAccessToken: nil,
	})
	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "trunk access denied"), err)
}

func TestController_SetLockState_trunkAccessToken_TrunkAccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	mockOperations.EXPECT().SetLockStateTrunkAccessToken(ctx, model.LOCKED, testdata.VinCar, testdata.TrunkAccessToken).Return(rentalErrors.ErrTrunkAccessDenied)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	token := testdata.TrunkAccessToken
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       nil,
		TrunkAccessToken: &token,
	})
	assert.Equal(t, echo.NewHTTPError(http.StatusForbidden, "trunk access denied"), err)
}

func TestController_SetLockState_NoParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContext := mocks.NewMockContext(ctrl)

	mockOperations := mocks.NewMockIOperations(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       nil,
		TrunkAccessToken: nil,
	})
	assert.Equal(t, echo.NewHTTPError(http.StatusBadRequest, "either customerId or trunkAccessToken must be specified"), err)
}

func TestController_SetLockState_TooManyParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContext := mocks.NewMockContext(ctrl)

	mockOperations := mocks.NewMockIOperations(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	token := testdata.TrunkAccessToken
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       &exampleCustomerID,
		TrunkAccessToken: &token,
	})
	assert.Equal(t, echo.NewHTTPError(http.StatusBadRequest, "only one of customerId or trunkAccessToken can be specified"), err)
}

func TestController_SetLockState_customerId_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().SetLockStateCustomerId(ctx, model.LOCKED, testdata.VinCar, exampleCustomerID).Return(operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       &exampleCustomerID,
		TrunkAccessToken: nil,
	})
	assert.ErrorIs(t, operationsError, err)
}

func TestController_SetLockState_trunkAccessToken_OperationsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)
	mockContext.EXPECT().Bind(gomock.Any()).SetArg(0, model.LockStateObject{TrunkLockState: model.LOCKED}).Return(nil)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().SetLockStateTrunkAccessToken(ctx, model.LOCKED, testdata.VinCar, testdata.TrunkAccessToken).Return(operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	token := testdata.TrunkAccessToken
	err := controller.SetLockState(mockContext, testdata.VinCar, model.SetLockStateParams{
		CustomerId:       nil,
		TrunkAccessToken: &token,
	})
	assert.ErrorIs(t, operationsError, err)
}
