package api

import (
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/mocks"
	"RentalManagement/testdata"
	"context"
	"errors"
	carTypes "git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-cargotypes.git"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var exampleCustomerID = "M9hUnd8a"

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
	Active: false,
	Car:    &carBase1,
	Id:     "M9hUnd8a",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
	},
}

var rentalCustomerShort2 = model.Rental{
	Active: true,
	Car:    &carBase2,
	Id:     "P2zUdL3C",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 1, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
}

var customerRentalsShort = []model.Rental{rentalCustomerShort1, rentalCustomerShort2}

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

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockEchoContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
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

	assert.Equal(t, err, echo.NewHTTPError(http.StatusBadRequest, "startDate must be before endDate"))
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
	assert.Equal(t, err, echo.NewHTTPError(http.StatusNotFound, "car not found"))
}

func TestController_CreateRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)

	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).Return(nil)
	mockEchoContext.EXPECT().NoContent(http.StatusCreated)

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

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)

	operationsError := errors.New("operations error")
	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)
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
	mockOperations := mocks.NewMockIOperations(ctrl)

	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockEchoContext.EXPECT().Request().Return(request)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).
		Return(rentalErrors.ErrCarNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(currentTime)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})
	assert.Equal(t, err, echo.NewHTTPError(http.StatusNotFound, "car not found"))
}

func TestController_CreateRental_invalidTimePeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)
	mockTime := mocks.NewMockITimeProvider(ctrl)

	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, invalidTimePeriod).Return(nil)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, err, echo.NewHTTPError(http.StatusBadRequest, "startDate must be before endDate"))
}

func TestController_CreateRental_Past(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)
	mockTime := mocks.NewMockITimeProvider(ctrl)

	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod1900).Return(nil)
	mockTime.EXPECT().Now().Return(future)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, err, echo.NewHTTPError(http.StatusForbidden, "startDate must be in the future"))
}

func TestController_CreateRental_ConflictingRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	request, _ := http.NewRequestWithContext(ctx, "POST", "", nil)

	mockEchoContext := mocks.NewMockContext(ctrl)
	mockOperations := mocks.NewMockIOperations(ctrl)
	mockTime := mocks.NewMockITimeProvider(ctrl)

	mockEchoContext.EXPECT().Bind(gomock.Any()).SetArg(0, timePeriod).Return(nil)
	mockTime.EXPECT().Now().Return(currentTime)
	mockEchoContext.EXPECT().Request().Return(request)
	mockOperations.EXPECT().CreateRental(ctx, testdata.VinCar, exampleCustomerID, timePeriod).
		Return(rentalErrors.ErrConflictingRentalExists)

	controller := NewController(mockOperations, mockTime)
	err := controller.CreateRental(mockEchoContext, testdata.VinCar,
		model.CreateRentalParams{CustomerId: exampleCustomerID})

	assert.Equal(t, err, echo.NewHTTPError(http.StatusConflict, "conflicting rental exists"))
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

	request, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

	mockContext := mocks.NewMockContext(ctrl)
	mockContext.EXPECT().Request().Return(request)

	mockOperations := mocks.NewMockIOperations(ctrl)
	operationsError := errors.New("operations error")
	mockOperations.EXPECT().GetOverview(ctx, exampleCustomerID).Return(nil, operationsError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	controller := NewController(mockOperations, mockTime)
	err := controller.GetOverview(mockContext, model.GetOverviewParams{CustomerId: exampleCustomerID})
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

	assert.Equal(t, err, echo.NewHTTPError(http.StatusNotFound, "rentalId not found"))
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
