package mappers

import (
	"RentalManagement/infrastructure/database/entities"
	"RentalManagement/logic/model"
	"RentalManagement/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var carVin1 = "G1YZ23J9P58034278"
var carVin2 = "1GKLVNED8AJ200101"

var car1 = entities.Car{
	Vin: carVin1,
	Rentals: []entities.Rental{
		rental1,
		rental2,
	},
}

var car2 = entities.Car{
	Vin: carVin2,
	Rentals: []entities.Rental{
		rental3,
	},
}

var cars = []entities.Car{
	car1,
	car2,
}

var rental1 = entities.Rental{
	RentalId:   "rZ6IIwcD",
	CustomerId: "M9hUnd8a",
	RentalPeriod: entities.TimePeriod{
		StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
	},
	TrunkToken: nil,
}

var rental2 = entities.Rental{
	RentalId:   "8J7szB1d",
	CustomerId: "d9COw9vI",
	RentalPeriod: entities.TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
	TrunkToken: &entities.TrunkAccessToken{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
		},
	},
}

var rental3 = entities.Rental{
	RentalId:   "u8NbZuNa",
	CustomerId: "nM8nB6Zu",
	RentalPeriod: entities.TimePeriod{
		StartDate: time.Date(2020, 1, 9, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC),
	},
	TrunkToken: nil,
}

var rentalModel1Car1 = model.Rental{
	State:    model.ACTIVE,
	Car:      &model.Car{Vin: "G1YZ23J9P58034278"},
	Customer: &model.Customer{CustomerId: "M9hUnd8a"},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
	},
	Token: nil,
}

var rentalModel2Car1 = model.Rental{
	State:    model.UPCOMING,
	Car:      &model.Car{Vin: "G1YZ23J9P58034278"},
	Customer: &model.Customer{CustomerId: "d9COw9vI"},
	Id:       "8J7szB1d",
	RentalPeriod: model.TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
	Token: &model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
		},
	},
}

var rentalModel3Car2 = model.Rental{
	State:    model.EXPIRED,
	Car:      &model.Car{Vin: "1GKLVNED8AJ200101"},
	Customer: &model.Customer{CustomerId: "nM8nB6Zu"},
	Id:       "u8NbZuNa",
	RentalPeriod: model.TimePeriod{
		StartDate: time.Date(2020, 1, 9, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC),
	},
	Token: nil,
}

var rentalsModelCar1 = []model.Rental{rentalModel1Car1, rentalModel2Car1}
var rentalsModelAll = []model.Rental{rentalModel1Car1, rentalModel2Car1, rentalModel3Car2}
var vins = []model.Vin{carVin1, carVin2}

var currentTime = time.Date(2023, 2, 2, 3, 10, 12, 100, time.UTC)

func TestMapTimePeriodToDb(t *testing.T) {
	assert.Equal(t, rental2.RentalPeriod, MapTimePeriodToDb(&rentalModel2Car1.RentalPeriod))
}

func TestMapCarSliceToVinSlice(t *testing.T) {
	assert.Equal(t, vins, MapCarSliceToVinSlice(&cars))
}

func TestMapTokenToDb(t *testing.T) {
	assert.Equal(t, rental2.TrunkToken, MapTokenToDb(rentalModel2Car1.Token))
}

func TestMapTokenToDb_Nil(t *testing.T) {
	assert.Nil(t, MapTokenToDb(nil))
}

func TestMapCarFromDbToRentals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tp := mocks.NewMockITimeProvider(ctrl)
	tp.EXPECT().Now().Return(currentTime)
	tp.EXPECT().Now().Return(currentTime)

	assert.Equal(t, rentalsModelCar1, MapCarFromDbToRentals(&car1, tp))
}

func TestMapCarsFromDbToRentals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tp := mocks.NewMockITimeProvider(ctrl)
	tp.EXPECT().Now().Return(currentTime)
	tp.EXPECT().Now().Return(currentTime)
	tp.EXPECT().Now().Return(currentTime)

	assert.Equal(t, rentalsModelAll, MapCarsFromDbToRentals(&cars, tp))
}
