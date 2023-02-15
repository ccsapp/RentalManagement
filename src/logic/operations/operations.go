package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"context"
	"fmt"
	"net/http"
)

type operations struct {
	carClient car.ClientWithResponsesInterface
	crud      database.ICRUD
}

func NewOperations(carClient car.ClientWithResponsesInterface, crud database.ICRUD) IOperations {
	return &operations{
		carClient: carClient,
		crud:      crud,
	}
}

func (o *operations) GetAvailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.CarAvailable, error) {
	carsResponse, err := o.carClient.GetCarsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if carsResponse.ParsedVins == nil {
		return nil, fmt.Errorf("%w: unknown error (status code %d)", rentalErrors.ErrDomainAssertion,
			carsResponse.StatusCode())
	}

	allCars := carsResponse.ParsedVins
	unavailableCars, err := o.crud.GetUnavailableCars(ctx, timePeriod)

	if err != nil {
		return nil, err
	}

	unavailable := map[model.Vin]bool{}
	for _, vin := range *unavailableCars {
		unavailable[vin] = true
	}

	availableCars := make([]model.CarAvailable, 0, len(*allCars)-len(*unavailableCars))
	for _, vin := range *allCars {
		if !unavailable[vin] {
			availableCar, err := o.getAvailableCar(ctx, vin)
			if err != nil {
				return nil, err
			}
			availableCars = append(availableCars, *availableCar)
		}
	}
	return &availableCars, nil
}

func (o *operations) getAvailableCar(ctx context.Context, vin model.Vin) (*model.CarAvailable, error) {
	carResponse, err := o.carClient.GetCarWithResponse(ctx, vin)
	if err != nil {
		return nil, err
	}
	if carResponse.ParsedCar == nil {
		return nil, fmt.Errorf("%w: unknown car %s (maybe the domain service is a little bit forgetful?)",
			rentalErrors.ErrDomainAssertion, vin)
	}
	return car.MapToCarAvailable(carResponse.ParsedCar), nil
}

func (o *operations) CreateRental(ctx context.Context, vin model.Vin, customerID model.CustomerId, timePeriod model.TimePeriod) error {
	if err := o.ensureCarExists(ctx, vin); err != nil {
		return err
	}
	return o.crud.CreateRental(ctx, vin, customerID, timePeriod)
}

func (o *operations) ensureCarExists(ctx context.Context, vin model.Vin) error {
	carResponse, err := o.carClient.GetCarWithResponse(ctx, vin)
	if err != nil {
		return err
	}
	if carResponse.StatusCode() == http.StatusNotFound {
		return rentalErrors.ErrCarNotFound
	}
	if carResponse.ParsedCar == nil {
		return rentalErrors.ErrDomainAssertion
	}
	return nil
}

func (o *operations) GetCar(ctx context.Context, vin model.Vin) (*model.Car, error) {
	carResponse, err := o.carClient.GetCarWithResponse(ctx, vin)
	if err != nil {
		return nil, err
	}
	if carResponse.StatusCode() == http.StatusNotFound {
		return nil, rentalErrors.ErrCarNotFound
	}
	if carResponse.ParsedCar == nil {
		return nil, rentalErrors.ErrDomainAssertion
	}
	return car.MapToCarStatic(carResponse.ParsedCar), nil
}

func (o *operations) GetOverview(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error) {
	rentals, err := o.crud.GetRentalsOfCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	for i, rental := range *rentals {
		carResponse, err := o.carClient.GetCarWithResponse(ctx, rental.Car.Vin)
		if err != nil {
			return nil, err
		}
		statusCode := carResponse.StatusCode()
		if statusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: car %s from customer %s not in domain",
				rentalErrors.ErrDomainAssertion, rental.Car.Vin, customerID)
		}
		if carResponse.ParsedCar == nil {
			return nil, fmt.Errorf("%w: unknown error (domain code %d)",
				rentalErrors.ErrDomainAssertion, statusCode)
		}
		rental.Car = car.MapToCarBase(carResponse.ParsedCar)
		(*rentals)[i] = rental.ToRentalCustomerShort()
	}

	return rentals, nil
}

func (o *operations) GetRentalStatus(ctx context.Context, rentalId model.RentalId) (*model.Rental, error) {
	rental, err := o.crud.GetRental(ctx, rentalId)
	if err != nil {
		return nil, err
	}

	carResponse, err := o.carClient.GetCarWithResponse(ctx, rental.Car.Vin)
	if err != nil {
		return nil, err
	}

	statusCode := carResponse.StatusCode()
	if statusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: car %s with rentalId %s not in domain",
			rentalErrors.ErrDomainAssertion, rental.Car.Vin, rentalId)
	}
	if carResponse.ParsedCar == nil {
		return nil, fmt.Errorf("%w: unknown error (domain code %d)",
			rentalErrors.ErrDomainAssertion, statusCode)
	}

	rentalReturn := rental.ToRentalCustomer()
	if rentalReturn.Active {
		rentalReturn.Car = car.MapToCar(carResponse.ParsedCar)
	} else {
		rentalReturn.Car = car.MapToCarStatic(carResponse.ParsedCar)
	}

	return &rentalReturn, nil
}
