package entities

import (
	"RentalManagement/logic/model"
	"time"
)

// Rental A rental
type Rental struct {
	// RentalId Unique identification of a rental
	RentalId model.RentalId `bson:"_id"`

	// CustomerId Unique identification of a customer
	CustomerId model.CustomerId `bson:"customer"`

	// Car The car associated with this rental
	Car model.Vin `bson:"car"`

	// RentalPeriod The time the rental is active, that is the time the car is rented
	RentalPeriod TimePeriod `bson:"rentalPeriod"`

	// TrunkToken Trunk access token
	TrunkToken *TrunkAccessToken `bson:"trunkToken,omitempty"`
}

// TimePeriod A time period
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
