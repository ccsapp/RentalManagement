package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/util"
	"context"
	"errors"
	"fmt"
	carTypes "github.com/ccsapp/cargotypes"
	"net/http"
)

type operations struct {
	carClient    car.ClientWithResponsesInterface
	crud         database.ICRUD
	timeProvider util.ITimeProvider
}

func NewOperations(carClient car.ClientWithResponsesInterface, crud database.ICRUD,
	timeProvider util.ITimeProvider) IOperations {
	return &operations{
		carClient:    carClient,
		crud:         crud,
		timeProvider: timeProvider,
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

func (o *operations) GetNextRental(ctx context.Context, vin model.Vin) (*model.Rental, error) {
	if err := o.ensureCarExists(ctx, vin); err != nil {
		return nil, err
	}
	rental, err := o.crud.GetNextRental(ctx, vin)
	if err != nil {
		return nil, err
	}
	if rental == nil {
		return nil, nil
	}
	nextRental := rental.ToRentalFleetManager()
	return &nextRental, nil
}

func (o *operations) CreateRental(ctx context.Context, vin model.Vin, customerID model.CustomerId,
	timePeriod model.TimePeriod) error {
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

func (o *operations) GrantTrunkAccess(ctx context.Context, rentalId model.RentalId, timePeriod model.TimePeriod) (
	*model.TrunkAccess, error) {

	trunkAccess := model.TrunkAccess{
		Token:          util.GenerateRandomString(24),
		ValidityPeriod: timePeriod,
	}

	createdToken, err := o.crud.SetTrunkToken(ctx, rentalId, trunkAccess)
	if errors.Is(err, database.OptimisticLockingError) {
		return nil, rentalErrors.ErrResourceConflict
	}
	if err != nil {
		return nil, err
	}

	return createdToken, nil
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
	if rentalReturn.State == model.ACTIVE {
		rentalReturn.Car = car.MapToCar(carResponse.ParsedCar)
	} else {
		rentalReturn.Car = car.MapToCarStatic(carResponse.ParsedCar)
	}

	return &rentalReturn, nil
}

func (o *operations) GetLockState(ctx context.Context, vin model.Vin, token model.TrunkAccessToken) (*model.LockState,
	error) {

	access, err := o.crud.GetTrunkAccess(ctx, vin, token)
	if err != nil {
		return nil, err
	}
	now := o.timeProvider.Now()
	if access.ValidityPeriod.EndDate.Before(now) || access.ValidityPeriod.StartDate.After(now) {
		return nil, rentalErrors.ErrTrunkAccessDenied
	}

	carResponse, err := o.carClient.GetCarWithResponse(ctx, vin)
	if err != nil {
		return nil, err
	}

	statusCode := carResponse.StatusCode()
	if statusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: car %s not in domain",
			rentalErrors.ErrDomainAssertion, vin)
	}
	if carResponse.ParsedCar == nil {
		return nil, fmt.Errorf("%w: unknown error (domain code %d)",
			rentalErrors.ErrDomainAssertion, statusCode)
	}
	lockState := car.MapToCar(carResponse.ParsedCar).DynamicData.TrunkLockState

	return &lockState, nil
}

func (o *operations) SetLockStateCustomerId(ctx context.Context, lockState model.LockState, vin model.Vin,
	customerId model.CustomerId) error {

	rental, err := o.crud.GetNextRental(ctx, vin)
	if err != nil {
		return err
	}
	if rental == nil || rental.Customer.CustomerId != customerId || rental.State != model.ACTIVE {
		return rentalErrors.ErrTrunkAccessDenied
	}

	return o.setLockState(ctx, lockState, vin)
}

func (o *operations) SetLockStateTrunkAccessToken(ctx context.Context, lockState model.LockState, vin model.Vin,
	token model.TrunkAccessToken) error {

	access, err := o.crud.GetTrunkAccess(ctx, vin, token)
	if err != nil {
		return err
	}
	now := o.timeProvider.Now()
	if access.ValidityPeriod.EndDate.Before(now) || access.ValidityPeriod.StartDate.After(now) {
		return rentalErrors.ErrTrunkAccessDenied
	}

	return o.setLockState(ctx, lockState, vin)
}

func (o *operations) setLockState(ctx context.Context, lockState model.LockState, vin model.Vin) error {
	response, err := o.carClient.ChangeTrunkLockStateWithResponse(ctx, vin, carTypes.DynamicDataLockState(lockState))
	if err != nil {
		return err
	}
	if response.HTTPResponse.StatusCode != http.StatusNoContent {
		return rentalErrors.ErrDomainAssertion
	}
	return nil
}
