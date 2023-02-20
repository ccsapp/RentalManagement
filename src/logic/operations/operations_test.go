package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
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

func TestOperations_GetOverview_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockCar := mocks.NewMockClientWithResponsesInterface(ctrl)
	mockCar.EXPECT().GetCarWithResponse(ctx, vin2).Return(&car.GetCarResponse{ParsedCar: &domainCar}, nil)

	mockCrud := mocks.NewMockICRUD(ctrl)
	mockCrud.EXPECT().GetRentalsOfCustomer(ctx, exampleCustomerID).Return(&[]model.Rental{rentalCrud}, nil)

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
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

	operations := NewOperations(mockCar, mockCrud)
	trunkAccess, err := operations.GrantTrunkAccess(ctx, "rentalId", timePeriod)

	assert.ErrorIs(t, err, rentalErrors.ErrResourceConflict)
	assert.Nil(t, trunkAccess)
}
