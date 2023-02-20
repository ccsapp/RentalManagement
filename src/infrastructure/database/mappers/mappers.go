package mappers

import (
	"RentalManagement/infrastructure/database/entities"
	"RentalManagement/logic/model"
	"RentalManagement/util"
	"time"
)

func MapTimePeriodToDb(period *model.TimePeriod) entities.TimePeriod {
	return entities.TimePeriod{
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
	}
}

func mapTimePeriodFromDb(period *entities.TimePeriod) model.TimePeriod {
	return model.TimePeriod{
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
	}
}

func MapCarSliceToVinSlice(cars *[]entities.Car) []model.Vin {
	vins := make([]model.Vin, len(*cars))
	for i, car := range *cars {
		vins[i] = car.Vin
	}
	return vins
}

func mapTokenFromDb(token *entities.TrunkAccessToken) *model.TrunkAccess {
	if token == nil {
		return nil
	}
	return &model.TrunkAccess{
		Token:          token.Token,
		ValidityPeriod: mapTimePeriodFromDb(&token.ValidityPeriod),
	}
}

func MapTokenToDb(token *model.TrunkAccess) *entities.TrunkAccessToken {
	if token == nil {
		return nil
	}
	return &entities.TrunkAccessToken{
		Token:          token.Token,
		ValidityPeriod: MapTimePeriodToDb(&token.ValidityPeriod),
	}
}

func getState(period *entities.TimePeriod, currentTime time.Time) model.State {
	if !period.StartDate.Before(currentTime) {
		return model.UPCOMING
	}
	if period.EndDate.After(currentTime) {
		return model.ACTIVE
	}
	return model.EXPIRED
}

// mapRentalFromDb only sets the VIN of the car
func mapRentalFromDb(rental *entities.Rental, vin model.Vin, timeProvider util.ITimeProvider) model.Rental {
	return model.Rental{
		State:        getState(&rental.RentalPeriod, timeProvider.Now()),
		Car:          &model.Car{Vin: vin},
		Customer:     &model.Customer{CustomerId: rental.CustomerId},
		Id:           rental.RentalId,
		RentalPeriod: mapTimePeriodFromDb(&rental.RentalPeriod),
		Token:        mapTokenFromDb(rental.TrunkToken),
	}
}

func MapCarFromDbToRentals(car *entities.Car, timeProvider util.ITimeProvider) []model.Rental {
	rentals := make([]model.Rental, len(car.Rentals))
	for i, rental := range car.Rentals {
		rentals[i] = mapRentalFromDb(&rental, car.Vin, timeProvider)
	}
	return rentals
}

func MapCarsFromDbToRentals(cars *[]entities.Car, timeProvider util.ITimeProvider) []model.Rental {
	rentals := make([]model.Rental, 0)
	for _, car := range *cars {
		rentals = append(rentals, MapCarFromDbToRentals(&car, timeProvider)...)
	}
	return rentals
}
