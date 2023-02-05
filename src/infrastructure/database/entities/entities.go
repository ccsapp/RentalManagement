package entities

import (
	"RentalManagement/logic/model"
	"time"
)

type Car struct {
	// Vin Vehicle Identification Number
	Vin model.Vin `bson:"_id"`

	// Rentals all rentals for this car
	Rentals []Rental `bson:"rentals"`
}

type Rental struct {
	// RentalId Unique identification of a rental
	RentalId model.RentalId `bson:"rentalId"`

	// CustomerId Unique identification of a customer
	CustomerId model.CustomerId `bson:"customer"`

	// RentalPeriod The time the rental is active, that is the time the car is rented
	RentalPeriod TimePeriod `bson:"rentalPeriod"`

	// TrunkToken Trunk access token
	TrunkToken *TrunkAccessToken `bson:"trunkToken,omitempty"`
}

type TimePeriod struct {
	// StartDate Beginning of the time period
	StartDate time.Time `bson:"startDate"`

	// EndDate End of the time period
	EndDate time.Time `bson:"endDate"`
}

// TrunkAccessToken Trunk access token with time
type TrunkAccessToken struct {
	// Token Trunk access token
	Token model.TrunkAccessToken `bson:"token"`

	// ValidityPeriod the time the token is valid
	ValidityPeriod TimePeriod `bson:"validityPeriod"`
}
