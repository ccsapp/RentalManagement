package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/mocks"
	"context"
	"errors"
	carTypes "github.com/ccsapp/cargotypes"
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

var timePeriod2 = model.TimePeriod{
	StartDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
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
	Vin: vin2,
}

var rentalCrud = model.Rental{
	State:    model.ACTIVE,
	Car:      &model.Car{Vin: domainCar.Vin},
	Customer: &model.Customer{CustomerId: exampleCustomerID},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCrudExpired = model.Rental{
	State:    model.EXPIRED,
	Car:      &model.Car{Vin: domainCar.Vin},
	Customer: &model.Customer{CustomerId: exampleCustomerID},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(1900, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCrudUpcoming = model.Rental{
	State:    model.UPCOMING,
	Car:      &model.Car{Vin: domainCar.Vin},
	Customer: &model.Customer{CustomerId: exampleCustomerID},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(1900, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCustomerShort = model.Rental{
	State: model.ACTIVE,
	Car:   &model.Car{Vin: domainCar.Vin, Brand: "Tesla", Model: "Model X"},
	Id:    "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
}

var rentalCustomerActive = model.Rental{
	State: model.ACTIVE,
	Car: &model.Car{
		Brand: "Tesla",
		DynamicData: &model.DynamicData{
			DoorsLockState:      model.UNLOCKED,
			EngineState:         model.OFF,
			FuelLevelPercentage: 20,
			Position: carTypes.DynamicDataPosition{
				Latitude:  49,
				Longitude: 8,
			},
			TrunkLockState: model.LOCKED,
		},
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
			Weight:        2000,
		},
		Vin: vin2,
	},
	Id: "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCustomerExpired = model.Rental{
	State: model.EXPIRED,
	Car: &model.Car{
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
			Weight:        2000,
		},
		Vin: vin2,
	},
	Id: "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(1900, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCustomerUpcoming = model.Rental{
	State: model.UPCOMING,
	Car: &model.Car{
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
			Weight:        2000,
		},
		Vin: vin2,
	},
	Id: "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(1900, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(1900, 3, 3, 1, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
	},
}

var rentalFleetManager = model.Rental{
	State:    model.ACTIVE,
	Customer: &model.Customer{CustomerId: exampleCustomerID},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		EndDate:   time.Date(2023, 4, 2, 3, 0, 0, 0, time.UTC),
		StartDate: time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
	},
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retCar, err := operations.GetCar(ctx, vin1)

	assert.ErrorIs(t, err, rentalErrors.ErrCarNotFound)
	assert.Nil(t, retCar)
}

func TestOperations_GetNextRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)
	mockCrud.EXPECT().GetNextRental(ctx, vin1).Return(&rentalCrud, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retRental, err := operations.GetNextRental(ctx, vin1)

	assert.Nil(t, err)
	assert.Equal(t, &rentalFleetManager, retRental)
}

func TestOperations_GetNextRental_noRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)
	mockCrud.EXPECT().GetNextRental(ctx, vin1).Return(nil, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retRental, err := operations.GetNextRental(ctx, vin1)

	assert.Nil(t, err)
	assert.Nil(t, retRental)
}

func TestOperations_GetNextRental_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	crudError := errors.New("crud error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCrud := mocks.NewMockICRUD(ctrl)

	mockCar.EXPECT().GetCarWithResponse(ctx, vin1).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)
	mockCrud.EXPECT().GetNextRental(ctx, vin1).Return(nil, crudError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retRental, err := operations.GetNextRental(ctx, vin1)

	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, retRental)
}

func TestOperations_GetNextRental_carNotFound(t *testing.T) {
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retRental, err := operations.GetNextRental(ctx, vin1)

	assert.ErrorIs(t, err, rentalErrors.ErrCarNotFound)
	assert.Nil(t, retRental)
}

func TestOperations_GetNextRental_unexpectedCarResponse(t *testing.T) {
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

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	retRental, err := operations.GetNextRental(ctx, vin1)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, retRental)
}

func TestOperations_GetOverview_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(&[]model.Rental{rentalCrud}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rentals, err := operations.GetOverview(ctx, exampleCustomerID)

	assert.Nil(t, err)
	assert.Equal(t, &[]model.Rental{rentalCustomerShort}, rentals)
}

func TestOperations_GetOverview_CrudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	crudError := errors.New("crud error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(nil, crudError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rentals, err := operations.GetOverview(ctx, exampleCustomerID)

	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, rentals)
}

func TestOperations_GetOverview_DomainError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	domainError := errors.New("domain error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(nil, domainError)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(&[]model.Rental{rentalCrud}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rentals, err := operations.GetOverview(ctx, exampleCustomerID)

	assert.ErrorIs(t, err, domainError)
	assert.Nil(t, rentals)
}

func TestOperations_GetOverview_CarNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(&[]model.Rental{rentalCrud}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rentals, err := operations.GetOverview(ctx, exampleCustomerID)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, rentals)
}

func TestOperations_GetOverview_UnknownDomainResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(&[]model.Rental{rentalCrud}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rentals, err := operations.GetOverview(ctx, exampleCustomerID)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, rentals)
}

func TestOperations_GetRentalStatus_success_active(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCrud.Id).Return(&rentalCrud, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCrud.Id)

	assert.Nil(t, err)
	assert.Equal(t, &rentalCustomerActive, rental)
}

func TestOperations_GetRentalStatus_success_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCrud.Id).Return(&rentalCrudExpired, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCrud.Id)

	assert.Nil(t, err)
	assert.Equal(t, &rentalCustomerExpired, rental)
}

func TestOperations_GetRentalStatus_success_Upcoming(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCrud.Id).Return(&rentalCrudUpcoming, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCrud.Id)

	assert.Nil(t, err)
	assert.Equal(t, &rentalCustomerUpcoming, rental)
}

func TestOperations_GetRentalStatus_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	crudError := errors.New("crud error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, "rentalId").Return(nil, crudError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, "rentalId")

	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, rental)
}

func TestOperations_GetRentalStatus_domainError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	domainError := errors.New("domain error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(nil, domainError)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCustomerShort.Id).Return(&rentalCrud, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCustomerShort.Id)

	assert.ErrorIs(t, err, domainError)
	assert.Nil(t, rental)
}

func TestOperations_GetRentalStatus_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCustomerShort.Id).Return(&rentalCrud, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCustomerShort.Id)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, rental)
}

func TestOperations_GetRentalStatus_UnknownDomainResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRental(ctx, rentalCustomerShort.Id).Return(&rentalCrud, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	rental, err := operations.GetRentalStatus(ctx, rentalCustomerShort.Id)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, rental)
}

func TestOperations_GrantTrunkAccess_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, trunkAccess model.TrunkAccess) (*model.TrunkAccess, error) {
			assert.Equal(t, timePeriod, trunkAccess.ValidityPeriod)
			return &model.TrunkAccess{Token: trunkAccess.Token, ValidityPeriod: timePeriod2}, nil
		},
	)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.Nil(t, err)
	assert.Equal(t, 24, len(trunkAccess.Token))
	assert.Equal(t, timePeriod2, trunkAccess.ValidityPeriod)
}

func TestOperations_GrantTrunkAccess_unknownRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).Return(nil, rentalErrors.ErrRentalNotFound)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotFound)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GrantTrunkAccess_unexpectedCrudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	crudError := errors.New("crud error")

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).Return(nil, crudError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GrantTrunkAccess_rentalExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).
		Return(nil, rentalErrors.ErrRentalNotActive)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotActive)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GrantTrunkAccess_rentalUpcoming(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).
		Return(nil, rentalErrors.ErrRentalNotActive)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotActive)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GrantTrunkAccess_rentalNotOverlapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).
		Return(nil, rentalErrors.ErrRentalNotOverlapping)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotOverlapping)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GrantTrunkAccess_resourceConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().SetTrunkToken(ctx, "rentalId", gomock.Any()).
		Return(nil, database.OptimisticLockingError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrResourceConflict)
	assert.Nil(t, trunkAccess)
}

func TestOperations_GetLockState_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)
	assert.Nil(t, err)
	assert.Equal(t, model.LOCKED, *lockState)
}

func TestOperations_GetLockState_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	crudError := errors.New("crud error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(nil, crudError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)

	assert.ErrorIs(t, err, crudError)
	assert.Nil(t, lockState)
}

func TestOperations_GetLockState_validInFutureTrunkAccessDeniedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrudUpcoming.Token.Token).Return(rentalCrudUpcoming.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(1900, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)
	assert.Equal(t, rentalErrors.ErrTrunkAccessDenied, err)
	assert.Nil(t, lockState)
}

func TestOperations_GetLockState_validInPastTrunkAccessDeniedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrudExpired.Token.Token).Return(rentalCrudExpired.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(3000, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)
	assert.Equal(t, rentalErrors.ErrTrunkAccessDenied, err)
	assert.Nil(t, lockState)
}

func TestOperations_GetLockState_unknownDomainResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, lockState)
}

func TestOperations_GetLockState_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)

	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
	assert.Nil(t, lockState)
}

func TestOperations_GetLockState_domainError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	domainError := errors.New("domain error")

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(nil, domainError)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	lockState, err := operations.GetLockState(ctx, vin2, rentalCrud.Token.Token)

	assert.ErrorIs(t, err, domainError)
	assert.Nil(t, lockState)
}

func TestOperations_SetLockStateCustomerId_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrud, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNoContent,
		},
	}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.Nil(t, err)
}

func TestOperations_SetLockStateCustomerId_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	crudError := errors.New("crud error")
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(nil, crudError)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, crudError)
}

func TestOperations_SetLockStateCustomerId_upcomingRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrudUpcoming, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
}

func TestOperations_SetLockStateCustomerId_wrongCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrud, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, "wrong customer")
	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
}

func TestOperations_SetLockStateCustomerId_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrud, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
}

func TestOperations_SetLockStateCustomerId_noActiveOrUpcomingRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(nil, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
}

func TestOperations_SetLockStateCustomerId_domainError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrud, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	domainError := errors.New("domain error")
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(nil, domainError)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, domainError)
}

func TestOperations_SetLockStateCustomerId_unknownDomainResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetNextRental(ctx, vin2).Return(&rentalCrud, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateCustomerId(ctx, model.LOCKED, vin2, exampleCustomerID)
	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
}

func TestOperations_SetLockStateTrunkAccessToken_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNoContent,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.Nil(t, err)
}

func TestOperations_SetLockStateTrunkAccessToken_crudError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	crudError := errors.New("crud error")
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(nil, crudError)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, crudError)
}

func TestOperations_SetLockStateTrunkAccessToken_validInFutureTrunkAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrudUpcoming.Token.Token).Return(rentalCrudUpcoming.Token, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(1900, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
}

func TestOperations_SetLockStateTrunkAccessToken_validInPastTrunkAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrudExpired.Token.Token).Return(rentalCrudExpired.Token, nil)

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2100, 3, 2, 5, 0, 0, 0, time.UTC))

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
}

func TestOperations_SetLockStateTrunkAccessToken_carNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
}

func TestOperations_SetLockStateTrunkAccessToken_domainError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	domainError := errors.New("domain error")
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(nil, domainError)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, domainError)
}

func TestOperations_SetLockStateTrunkAccessToken_unknownDomainResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetTrunkAccess(ctx, vin2, rentalCrud.Token.Token).Return(rentalCrud.Token, nil)

	mockTime := mocks.NewMockITimeProvider(ctrl)
	mockTime.EXPECT().Now().Return(time.Date(2023, 3, 2, 5, 0, 0, 0, time.UTC))

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().ChangeTrunkLockStateWithResponse(ctx, vin2,
		carTypes.DynamicDataLockState(model.LOCKED)).Return(&car.ChangeTrunkLockStateResponse{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusTeapot,
		},
	}, nil)

	operations := NewOperations(mockCar, mockCrud, mockTime)
	err := operations.SetLockStateTrunkAccessToken(ctx, model.LOCKED, vin2, rentalCrud.Token.Token)
	assert.ErrorIs(t, err, rentalErrors.ErrDomainAssertion)
}
