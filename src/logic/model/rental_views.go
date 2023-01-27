package model

// ToRentalCustomer selects Active, Car, Id, RentalPeriod and Token. Customer is omitted.
func (r *Rental) ToRentalCustomer() Rental {
	return Rental{
		Active:       r.Active,
		Car:          r.Car,
		Id:           r.Id,
		Customer:     nil,
		RentalPeriod: r.RentalPeriod,
		Token:        r.Token,
	}
}

// ToRentalFleetManager selects Active, Id, Customer and RentalPeriod. Car and Token are omitted.
func (r *Rental) ToRentalFleetManager() Rental {
	return Rental{
		Active:       r.Active,
		Car:          nil,
		Id:           r.Id,
		Customer:     r.Customer,
		RentalPeriod: r.RentalPeriod,
		Token:        nil,
	}
}
