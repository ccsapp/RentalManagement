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

func MapRentalSliceToVinSlice(rentals *[]entities.Rental) []model.Vin {
	vins := make([]model.Vin, len(*rentals))
	for i, rental := range *rentals {
		vins[i] = rental.Car
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

func isActive(period *entities.TimePeriod, currentTime time.Time) bool {
	return period.StartDate.Before(currentTime) && period.EndDate.After(currentTime)
}

func MapRentalFromDb(rental *entities.Rental, timeProvider util.ITimeProvider) model.Rental {
	return model.Rental{
		Active:       isActive(&rental.RentalPeriod, timeProvider.Now()),
		Car:          &model.Car{Vin: rental.Car},
		Customer:     &model.Customer{CustomerId: rental.CustomerId},
		Id:           rental.RentalId,
		RentalPeriod: mapTimePeriodFromDb(&rental.RentalPeriod),
		Token:        mapTokenFromDb(rental.TrunkToken),
	}
}

func MapRentalSliceFromDb(rentals *[]entities.Rental, timeProvider util.ITimeProvider) []model.Rental {
	modelRentals := make([]model.Rental, len(*rentals))
	for i, rental := range *rentals {
		modelRentals[i] = MapRentalFromDb(&rental, timeProvider)
	}
	return modelRentals
}
