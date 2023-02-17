// Package model provides the types of the exposed openapi HTTP API.
package model

import "time"

// Defines values for DynamicDataEngineState.
const (
	OFF DynamicDataEngineState = "OFF"
	ON  DynamicDataEngineState = "ON"
)

// Defines values for LockState.
const (
	LOCKED   LockState = "LOCKED"
	UNLOCKED LockState = "UNLOCKED"
)

// Defines values for TechnicalSpecificationFuel.
const (
	DIESEL       TechnicalSpecificationFuel = "DIESEL"
	ELECTRIC     TechnicalSpecificationFuel = "ELECTRIC"
	HYBRIDDIESEL TechnicalSpecificationFuel = "HYBRID_DIESEL"
	HYBRIDPETROL TechnicalSpecificationFuel = "HYBRID_PETROL"
	PETROL       TechnicalSpecificationFuel = "PETROL"
)

// Defines values for TechnicalSpecificationTransmission.
const (
	AUTOMATIC TechnicalSpecificationTransmission = "AUTOMATIC"
	MANUAL    TechnicalSpecificationTransmission = "MANUAL"
)

// Car defines model for car.
type Car struct {
	// Brand Data that specifies the brand name of the manufacturer
	Brand string `json:"brand"`

	// DynamicData Data that changes during a car's operation
	DynamicData *DynamicData `json:"dynamicData,omitempty"`

	// Model Data that specifies the particular type of car
	Model                  string                  `json:"model"`
	TechnicalSpecification *TechnicalSpecification `json:"technicalSpecification,omitempty"`

	// Vin A Vehicle Identification Number (VIN) which uniquely identifies a car
	Vin Vin `json:"vin"`
}

// CarAvailable defines model for carAvailable.
type CarAvailable struct {
	// Brand Data that specifies the brand name of the manufacturer
	Brand string `json:"brand"`

	// Model Data that specifies the particular type of car
	Model string `json:"model"`

	// NumberOfSeats Data that defines the number of seats that are built into a car
	NumberOfSeats int `json:"numberOfSeats"`

	// Vin A Vehicle Identification Number (VIN) which uniquely identifies a car
	Vin Vin `json:"vin"`
}

// Customer A customer
type Customer struct {
	// CustomerId Unique identification of a customer
	CustomerId CustomerId `json:"customerId"`
}

// CustomerId Unique identification of a customer
type CustomerId = string

// DynamicData Data that changes during a car's operation
type DynamicData struct {
	// DoorsLockState Data that specifies whether an object is locked or unlocked
	DoorsLockState LockState              `json:"doorsLockState"`
	EngineState    DynamicDataEngineState `json:"engineState"`

	// FuelLevelPercentage Data that specifies the relation of remaining fuelCapacity to the maximum fuelCapacity in percentage
	FuelLevelPercentage int `json:"fuelLevelPercentage"`

	// Position Data that specifies the GeoCoordinate of a car
	Position struct {
		// Latitude Data that specifies the distance from the equator
		Latitude float32 `json:"latitude"`

		// Longitude Data that specifies the distance east or west from a line (meridian) passing through Greenwich
		Longitude float32 `json:"longitude"`
	} `json:"position"`

	// TrunkLockState Data that specifies whether an object is locked or unlocked
	TrunkLockState LockState `json:"trunkLockState"`
}

// DynamicDataEngineState defines model for DynamicData.EngineState.
type DynamicDataEngineState string

// LockState Data that specifies whether an object is locked or unlocked
type LockState string

// LockStateObject An object containing the trunk lock state
type LockStateObject struct {
	// TrunkLockState Data that specifies whether an object is locked or unlocked
	TrunkLockState LockState `json:"trunkLockState"`
}

// Rental defines a model for rentals.
type Rental struct {
	// Active Describes whether this rental is active
	Active bool `json:"active"`

	// Car The rented car
	Car *Car `json:"car,omitempty"`

	// Id Unique identification of a rental
	Id RentalId `json:"id"`

	// Customer The renting customer
	Customer *Customer `json:"customer,omitempty"`

	// RentalPeriod A period of time
	RentalPeriod TimePeriod `json:"rentalPeriod"`

	// Token Trunk access token with time
	Token *TrunkAccess `json:"token,omitempty"`
}

// RentalId Unique identification of a rental
type RentalId = string

// TechnicalSpecification defines model for technicalSpecification.
type TechnicalSpecification struct {
	// Color Data on the description of the paint job of a car
	Color string `json:"color"`

	// Consumption Data that specifies the amount of fuel consumed during car operation in units per 100 kilometers
	Consumption struct {
		// City Data that specifies the amount of fuel that is consumed when driving within the city in: kW/100km or l/100km
		City float32 `json:"city"`

		// Combined Data that specifies the combined amount of fuel that is consumed in: kW / 100 km or l / 100 km
		Combined float32 `json:"combined"`

		// Overland Data that specifies the amount of fuel that is consumed when driving outside a city in: kW/100km or l/100km
		Overland float32 `json:"overland"`
	} `json:"consumption"`

	// Emissions Data that specifies the CO2 emitted by a car during operation in gram per kilometer
	Emissions struct {
		// City Data that specifies the amount of emissions when driving within the city in: g CO2 / km
		City float32 `json:"city"`

		// Combined Data that specifies the combined amount of emissions in: g CO2 / km. The combination is done by the manufacturer according to an industry-specific standard
		Combined float32 `json:"combined"`

		// Overland Data that specifies the amount of emissions when driving outside a city in: g CO2 / km
		Overland float32 `json:"overland"`
	} `json:"emissions"`

	// Engine A physical unit that converts fuel into movement
	Engine struct {
		// Power Data on the power the engine can provide in kW
		Power int `json:"power"`

		// Type Data that contains the manufacturer-given type description of the engine
		Type string `json:"type"`
	} `json:"engine"`

	// Fuel Data that defines the source of energy that powers the car
	Fuel TechnicalSpecificationFuel `json:"fuel"`

	// FuelCapacity Data that specifies the amount of fuel that can be carried with the car
	FuelCapacity string `json:"fuelCapacity"`

	// NumberOfDoors Data that defines the number of doors that are built into a car
	NumberOfDoors int `json:"numberOfDoors"`

	// NumberOfSeats Data that defines the number of seats that are built into a car
	NumberOfSeats int `json:"numberOfSeats"`

	// Transmission A physical unit responsible for managing the conversion rate of the engine (can be automated or manually operated)
	Transmission TechnicalSpecificationTransmission `json:"transmission"`

	// TrunkVolume Data on the physical volume of the trunk in liters
	TrunkVolume int `json:"trunkVolume"`

	// Weight Data that specifies the total weight of a car when empty in kilograms (kg)
	Weight int `json:"weight"`
}

// TechnicalSpecificationFuel Data that defines the source of energy that powers the car
type TechnicalSpecificationFuel string

// TechnicalSpecificationTransmission A physical unit responsible for managing the conversion rate of the engine (can be automated or manually operated)
type TechnicalSpecificationTransmission string

// TimePeriod A period of time
type TimePeriod struct {
	// StartDate start of the time period
	StartDate time.Time `json:"startDate"`

	// EndDate end of the time period
	EndDate time.Time `json:"endDate"`
}

// RestrictTo restricts the time period to the given other time period
// and returns the result. If the time periods do not overlap, nil is returned.
func (tp *TimePeriod) RestrictTo(other *TimePeriod) *TimePeriod {
	outPeriod := &TimePeriod{
		StartDate: max(tp.StartDate, other.StartDate),
		EndDate:   min(tp.EndDate, other.EndDate),
	}

	if !outPeriod.StartDate.Before(outPeriod.EndDate) {
		return nil
	}

	return outPeriod
}

func min(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// TrunkAccess Trunk access token with time
type TrunkAccess struct {
	// Token Trunk access token
	Token TrunkAccessToken `json:"token"`

	// ValidityPeriod A period of time
	ValidityPeriod TimePeriod `json:"validityPeriod"`
}

// TrunkAccessToken Trunk access token
type TrunkAccessToken = string

// Vin A Vehicle Identification Number (VIN) which uniquely identifies a car
type Vin = string

// CustomerIdOptionalParam Unique identification of a customer
type CustomerIdOptionalParam = CustomerId

// CustomerIdParam Unique identification of a customer
type CustomerIdParam = CustomerId

// RentalIdParam Unique identification of a rental
type RentalIdParam = RentalId

// TrunkAccessTokenOptionalParam Trunk access token
type TrunkAccessTokenOptionalParam = TrunkAccessToken

// TrunkAccessTokenParam Trunk access token
type TrunkAccessTokenParam = TrunkAccessToken

// VinParam A Vehicle Identification Number (VIN) which uniquely identifies a car
type VinParam = Vin

// GetAvailableCarsParams defines parameters for GetAvailableCars.
type GetAvailableCarsParams struct {
	TimePeriod TimePeriod `form:"timePeriod" json:"timePeriod"`
}

// CreateRentalParams defines parameters for CreateRental.
type CreateRentalParams struct {
	// CustomerId Unique identification of a customer
	CustomerId CustomerIdParam `form:"customerId" json:"customerId"`
}

// GetLockStateParams defines parameters for GetLockState.
type GetLockStateParams struct {
	// TrunkAccessToken A trunk access token
	TrunkAccessToken TrunkAccessTokenParam `form:"trunkAccessToken" json:"trunkAccessToken"`
}

// SetLockStateParams defines parameters for SetLockState.
type SetLockStateParams struct {
	// CustomerId Unique identification of a customer
	CustomerId *CustomerIdOptionalParam `form:"customerId,omitempty" json:"customerId,omitempty"`

	// TrunkAccessToken A trunk access token
	TrunkAccessToken *TrunkAccessTokenOptionalParam `form:"trunkAccessToken,omitempty" json:"trunkAccessToken,omitempty"`
}

// GetOverviewParams defines parameters for GetOverview.
type GetOverviewParams struct {
	// CustomerId Unique identification of a customer
	CustomerId CustomerIdParam `form:"customerId" json:"customerId"`
}
