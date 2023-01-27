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

var rental1 = entities.Rental{
	RentalId:   "rZ6IIwcD",
	CustomerId: "M9hUnd8a",
	Car:        "G1YZ23J9P58034278",
	RentalPeriod: entities.TimePeriod{
		StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
	},
	TrunkToken: nil,
}

var rentalModel1 = model.Rental{
	Active:   true,
	Car:      &model.Car{Vin: "G1YZ23J9P58034278"},
	Customer: &model.Customer{CustomerId: "M9hUnd8a"},
	Id:       "rZ6IIwcD",
	RentalPeriod: model.TimePeriod{
		StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
	},
	Token: nil,
}

var rental2 = entities.Rental{
	RentalId:   "8J7szB1d",
	CustomerId: "d9COw9vI",
	Car:        "1GKLVNED8AJ200101",
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

var rentalModel2 = model.Rental{
	Active:   false,
	Car:      &model.Car{Vin: "1GKLVNED8AJ200101"},
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

var rentalsModel = []model.Rental{rentalModel1, rentalModel2}
var rentals = []entities.Rental{rental1, rental2}
var vins = []model.Vin{rental1.Car, rental2.Car}

var currentTime = time.Date(2023, 2, 2, 3, 10, 12, 100, time.UTC)

func TestMapTimePeriodToDb(t *testing.T) {
	assert.Equal(t, rental2.RentalPeriod, MapTimePeriodToDb(&rentalModel2.RentalPeriod))
}

func TestMapRentalSliceToVinSlice(t *testing.T) {
	assert.Equal(t, vins, MapRentalSliceToVinSlice(&rentals))
}

func TestMapTokenToDb(t *testing.T) {
	assert.Equal(t, rental2.TrunkToken, MapTokenToDb(rentalModel2.Token))
}

func TestMapTokenToDb_Nil(t *testing.T) {
	assert.Nil(t, MapTokenToDb(nil))
}

func TestMapRentalFromDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tp := mocks.NewMockITimeProvider(ctrl)
	tp.EXPECT().Now().Return(currentTime)
	tp.EXPECT().Now().Return(currentTime)

	assert.Equal(t, rentalModel1, MapRentalFromDb(&rental1, tp))
	assert.Equal(t, rentalModel2, MapRentalFromDb(&rental2, tp))
}

func TestMapRentalSliceFromDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tp := mocks.NewMockITimeProvider(ctrl)
	tp.EXPECT().Now().Return(currentTime)
	tp.EXPECT().Now().Return(currentTime)

	assert.Equal(t, rentalsModel, MapRentalSliceFromDb(&rentals, tp))
}
