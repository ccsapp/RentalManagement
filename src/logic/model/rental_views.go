package model

// ToRentalCustomer selects State, Car, Id, RentalPeriod and Token. Customer is omitted.
func (r *Rental) ToRentalCustomer() Rental {
	return Rental{
		State:        r.State,
		Car:          r.Car,
		Id:           r.Id,
		Customer:     nil,
		RentalPeriod: r.RentalPeriod,
		Token:        r.Token,
	}
}

// ToRentalCustomerShort selects State, Car, Id and RentalPeriod. Customer and Token are omitted.
func (r *Rental) ToRentalCustomerShort() Rental {
	return Rental{
		State:        r.State,
		Car:          r.Car,
		Id:           r.Id,
		Customer:     nil,
		RentalPeriod: r.RentalPeriod,
		Token:        nil,
	}
}

// ToRentalFleetManager selects State, Id, Customer and RentalPeriod. Car and Token are omitted.
func (r *Rental) ToRentalFleetManager() Rental {
	return Rental{
		State:        r.State,
		Car:          nil,
		Id:           r.Id,
		Customer:     r.Customer,
		RentalPeriod: r.RentalPeriod,
		Token:        nil,
	}
}
