package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/mocks"
	"context"
	"errors"
	carTypes "git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-cargotypes.git"
	openapiTypes "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var exampleCustomerID = "34tfewss"

var timePeriod = model.TimePeriod{
	StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
}

const vin1 = "WVWAA71K08W201030"
const vin2 = "1FVNY5Y90HP312888"

var domainCar = carTypes.Car{
	Brand: "Tesla",
	DynamicData: carTypes.DynamicData{
		DoorsLockState:      carTypes.UNLOCKED,
		EngineState:         carTypes.OFF,
		FuelLevelPercentage: 20,
		Position: carTypes.DynamicDataPosition{
			Latitude:  49,
			Longitude: 8,
		},
		TrunkLockState: carTypes.LOCKED,
	},
	Model: "Model X",
	ProductionDate: openapiTypes.Date{
		Time: time.Date(2023, 12, 10, 7, 1, 3, 0, time.UTC),
	},
	TechnicalSpecification: carTypes.TechnicalSpecification{
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
		Fuel:          carTypes.HYBRIDDIESEL,
		FuelCapacity:  "54.0L;85.2kWh",
		NumberOfDoors: 4,
		NumberOfSeats: 5,
		Tire: carTypes.TechnicalSpecificationTire{
			Manufacturer: "Michelin",
			Type:         "180 CDI",
		},
		Transmission: carTypes.MANUAL,
		TrunkVolume:  1000,
		Weight:       2000,
	},
	Vin: "1FVNY5Y90HP312888",
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

var carAvailable = model.CarAvailable{
	Brand:         "Tesla",
	Model:         "Model X",
	NumberOfSeats: 5,
	Vin:           "1FVNY5Y90HP312888",
}

func TestOperations_GetAvailableCars_unexpectedCarResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarsWithResponse(ctx).Return(&car.GetCarsResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)

	operations := NewOperations(mockCar, mockCrud)
	ret, err := operations.GetAvailableCars(ctx, timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, ret)
}

func TestOperations_GetAvailableCars_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarsWithResponse(ctx).Return(&car.GetCarsResponse{ParsedVins: &[]carTypes.Vin{vin2, vin1}}, nil)

	crudError := errors.New("crud error")

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetUnavailableCars(ctx, timePeriod).Return(nil, crudError)

	operations := NewOperations(mockCar, mockCrud)
	ret, err := operations.GetAvailableCars(ctx, timePeriod)
	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, ret)
}

func TestOperations_GetAvailableCars_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarsWithResponse(ctx).Return(&car.GetCarsResponse{ParsedVins: &[]carTypes.Vin{vin2, vin1}}, nil)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetUnavailableCars(ctx, timePeriod).Return(&[]model.Vin{vin1}, nil)

	operations := NewOperations(mockCar, mockCrud)
	ret, err := operations.GetAvailableCars(ctx, timePeriod)
	assert.Nil(t, err)
	assert.Equal(t, &[]model.CarAvailable{carAvailable}, ret)
}

func TestOperations_CreateRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)
	mockCrud.EXPECT().CreateRental(ctx, vin1, exampleCustomerID, timePeriod).Return(nil)

	operations := NewOperations(mockCar, mockCrud)
	err := operations.CreateRental(ctx, vin1, exampleCustomerID, timePeriod)
	assert.Nil(t, err)
}

func TestOperations_CreateRental_unexpectedCarResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud)
	err := operations.CreateRental(ctx, vin1, exampleCustomerID, timePeriod)
	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
}

func TestOperations_CreateRental_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud)
	err := operations.CreateRental(ctx, vin1, exampleCustomerID, timePeriod)
	assert.ErrorIs(t, err, rentalErrors.ErrCarNotFound)
}

func TestOperations_CreateRental_conflictingRentalExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)
	mockCrud.EXPECT().CreateRental(ctx, vin1, exampleCustomerID, timePeriod).Return(rentalErrors.ErrConflictingRentalExists)

	operations := NewOperations(mockCar, mockCrud)
	err := operations.CreateRental(ctx, vin1, exampleCustomerID, timePeriod)
	assert.ErrorIs(t, err, rentalErrors.ErrConflictingRentalExists)
}

func TestOperations_GetCar_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	operations := NewOperations(mockCar, mockCrud)
	retCar, err := operations.GetCar(ctx, vin1)

	assert.Nil(t, err)
	assert.Equal(t, &staticCar, retCar)
}

func TestOperations_GetCar_unexpectedCarResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud)
	retCar, err := operations.GetCar(ctx, vin1)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, retCar)
}

func TestOperations_GetCar_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud)
	retCar, err := operations.GetCar(ctx, vin1)

	assert.ErrorIs(t, err, rentalErrors.ErrCarNotFound)
	assert.Nil(t, retCar)
}
