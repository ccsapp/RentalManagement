package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var rental = Rental{
	State:    ACTIVE,
	Car:      &Car{Vin: "G1YZ23J9P58034280"},
	Customer: &Customer{CustomerId: "d9COwOvI"},
	Id:       "rZ6I3weD",
	RentalPeriod: TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
	Token: &TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: TimePeriod{
			StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCustomer = Rental{
	State: ACTIVE,
	Car:   &Car{Vin: "G1YZ23J9P58034280"},
	Id:    "rZ6I3weD",
	RentalPeriod: TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
	Token: &TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: TimePeriod{
			StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
		},
	},
}

var rentalCustomerShort = Rental{
	State: ACTIVE,
	Car:   &Car{Vin: "G1YZ23J9P58034280"},
	Id:    "rZ6I3weD",
	RentalPeriod: TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
}

var rentalFleetManager = Rental{
	State:    ACTIVE,
	Customer: &Customer{CustomerId: "d9COwOvI"},
	Id:       "rZ6I3weD",
	RentalPeriod: TimePeriod{
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
	},
}

func TestRental_ToRentalCustomer(t *testing.T) {
	assert.Equal(t, rentalCustomer, rental.ToRentalCustomer())
}

func TestRental_ToRentalFleetManager(t *testing.T) {
	assert.Equal(t, rentalFleetManager, rental.ToRentalFleetManager())
}

func TestRental_ToRentalCustomerShort(t *testing.T) {
	assert.Equal(t, rentalCustomerShort, rental.ToRentalCustomerShort())
}
