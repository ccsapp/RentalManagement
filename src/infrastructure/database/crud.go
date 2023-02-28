package database

//go:generate mockgen -source=./crud.go -package=mocks -destination=../../mocks/mock_crud.go

import (
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/infrastructure/database/entities"
	"RentalManagement/infrastructure/database/mappers"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/util"
	"context"
	"errors"
)

const CollectionBaseName = "rentals"

var OptimisticLockingError = errors.New("optimistic locking failed")

// ICRUD is a high level database interface. It directly maps to the business logic and abstracts away the
// database entities and the database connection.
type ICRUD interface {
	GetUnavailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.Vin, error)
	CreateRental(ctx context.Context, vin model.Vin, customerId model.CustomerId, timePeriod model.TimePeriod) error
	GetRentalsOfCustomer(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error)

	// SetTrunkToken sets the trunk token of a rental.
	// Any old trunk token is overwritten.
	// The validity period of the trunk token is restricted to the validity period of the rental.
	// The resulting trunk token written to the database is returned (nil if any error occurred).
	// If the rental does not exist, rentalErrors.ErrRentalNotFound is returned.
	// If the rental is not active, rentalErrors.ErrRentalNotActive is returned.
	// If the rental is not active at any time during the validity period,
	// rentalErrors.ErrRentalNotOverlapping is returned.
	// This method uses optimistic locking for race condition safety.
	// If an optimistic locking error occurs, the method is retried up to 2 times.
	// If the optimistic locking error persists, OptimisticLockingError is returned.
	SetTrunkToken(ctx context.Context, rentalId model.RentalId,
		trunkAccess model.TrunkAccess) (*model.TrunkAccess, error)
	GetRental(ctx context.Context, rentalId model.RentalId) (*model.Rental, error)
	// GetNextRental returns the active or next upcoming rental of a car. If there is no next rental, nil is returned.
	GetNextRental(ctx context.Context, vin model.Vin) (*model.Rental, error)
	// GetTrunkAccess returns the trunk access token of a rental.
	// If token is not registered with the car with the provided vin, rentalErrors.ErrTrunkAccessDenied is returned.
	GetTrunkAccess(ctx context.Context, vin model.Vin, token model.TrunkAccessToken) (*model.TrunkAccess, error)
}

type crud struct {
	db           db.IConnection
	collection   string
	timeProvider util.ITimeProvider
}

func NewICRUD(db db.IConnection, config *db.Config, provider util.ITimeProvider) ICRUD {
	return &crud{
		db:           db,
		collection:   config.CollectionPrefix + CollectionBaseName,
		timeProvider: provider,
	}
}

func (c *crud) GetUnavailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.Vin, error) {
	var cars []entities.Car

	factory := c.db.GetFactory()

	// startDate <= timePeriod.EndDate AND endDate >= timePeriod.StartDate
	// this call creates cars that have only their VIN set
	err := c.db.FindMany(
		ctx,
		c.collection,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterAnd(
				factory.FilterLess("rentalPeriod.startDate", timePeriod.EndDate),
				factory.FilterGreater("rentalPeriod.endDate", timePeriod.StartDate),
			),
		),
		&db.Options{Projection: factory.ProjectionID()},
		&cars,
	)
	if err != nil {
		return nil, err
	}

	// extract VINs from the rentals, remove duplicates
	vins := mappers.MapCarSliceToVinSlice(&cars)

	return &vins, nil
}

func (c *crud) CreateRental(ctx context.Context, vin model.Vin, customerId model.CustomerId,
	timePeriod model.TimePeriod) error {

	factory := c.db.GetFactory()

	// create a new rental
	rental := entities.Rental{
		RentalId:     util.GenerateRandomString(8),
		CustomerId:   customerId,
		RentalPeriod: mappers.MapTimePeriodToDb(&timePeriod),
	}

	err := c.db.UpdateOne(
		ctx,
		c.collection,
		factory.FilterAnd(
			factory.FilterEqual("_id", vin),
			factory.FilterNot(
				factory.FilterElementMatch(
					"rentals",
					factory.FilterAnd(
						factory.FilterLess("rentalPeriod.startDate", timePeriod.EndDate),
						factory.FilterGreater("rentalPeriod.endDate", timePeriod.StartDate),
					),
				),
			),
		),
		factory.UpdatePush("rentals", rental),
		true,
	)

	// If the update failed because of a duplicate key error, there exists a car (because of the duplicate key error)
	// that has a conflicting rental (because upsert chose insert)
	if errors.Is(err, db.DuplicateKeyError) {
		return rentalErrors.ErrConflictingRentalExists
	}

	return err
}

func (c *crud) GetRentalsOfCustomer(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error) {
	var cars []entities.Car

	factory := c.db.GetFactory()

	err := c.db.Aggregate(
		ctx, c.collection, factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.customer", customerID),
			-1, //no limit
			nil,
		), &cars,
	)
	if err != nil {
		return nil, err
	}

	rentals := mappers.MapCarsFromDbToRentals(&cars, c.timeProvider)

	return &rentals, nil
}

func (c *crud) SetTrunkToken(ctx context.Context, rentalId model.RentalId,
	trunkAccess model.TrunkAccess) (*model.TrunkAccess, error) {

	var err error
	var returnedAccess *model.TrunkAccess

	// if an optimistic locking error occurs, try again (but only twice)
	for i := 0; i < 3; i++ {
		returnedAccess, err = c.trySetTrunkToken(ctx, rentalId, trunkAccess)

		if !errors.Is(err, OptimisticLockingError) {
			break
		}
	}

	return returnedAccess, err
}

func (c *crud) trySetTrunkToken(ctx context.Context, rentalId model.RentalId,
	trunkAccess model.TrunkAccess) (*model.TrunkAccess, error) {

	factory := c.db.GetFactory()

	var cars []entities.Car

	// fetch the rental with the given rentalId
	err := c.db.Aggregate(
		ctx,
		c.collection,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", rentalId),
			1, // limit to 1
			nil,
		),
		&cars,
	)

	if err != nil {
		return nil, err
	}

	if len(cars) == 0 {
		return nil, rentalErrors.ErrRentalNotFound
	}

	if len(cars) > 1 {
		panic("more than one car returned for a single rentalId")
	}

	if len(cars[0].Rentals) != 1 {
		panic("returned car has wrong number of rentals")
	}

	rentalEntity := cars[0].Rentals[0]
	rentalModel := mappers.MapCarFromDbToRentals(&cars[0], c.timeProvider)[0]

	if rentalModel.State != model.ACTIVE {
		return nil, rentalErrors.ErrRentalNotActive
	}

	restrictedValidityPeriod := trunkAccess.ValidityPeriod.RestrictTo(&rentalModel.RentalPeriod)
	if restrictedValidityPeriod == nil {
		return nil, rentalErrors.ErrRentalNotOverlapping
	}

	trunkAccess.ValidityPeriod = *restrictedValidityPeriod

	trunkTokenEntity := mappers.MapTokenToDb(&trunkAccess)

	// Optimistic Locking: If the rental changed in the meantime, the update will not do anything
	// (i.e. return NoDocumentsError)
	err = c.db.UpdateOne(
		ctx,
		c.collection,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(rentalEntity),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			*trunkTokenEntity,
		),
		false, // no upsert
	)

	if errors.Is(err, db.NoDocumentsError) {
		return nil, OptimisticLockingError
	}

	if err != nil {
		return nil, err
	}

	return &trunkAccess, nil
}

func (c *crud) GetRental(ctx context.Context, rentalId model.RentalId) (*model.Rental, error) {
	var cars []entities.Car

	factory := c.db.GetFactory()

	err := c.db.Aggregate(
		ctx, c.collection, factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", rentalId),
			1,
			nil,
		), &cars,
	)
	if err != nil {
		return nil, err
	}
	if len(cars) == 0 {
		return nil, rentalErrors.ErrRentalNotFound
	}

	rentals := mappers.MapCarFromDbToRentals(&cars[0], c.timeProvider)
	return &rentals[0], nil
}

func (c *crud) GetNextRental(ctx context.Context, vin model.Vin) (*model.Rental, error) {
	var cars []entities.Car

	factory := c.db.GetFactory()

	err := c.db.Aggregate(ctx, c.collection, factory.ArrayFilterAggregation(
		"rentals",
		factory.FilterAnd(
			factory.FilterEqual("_id", vin),
			factory.FilterGreater("rentals.rentalPeriod.endDate", c.timeProvider.Now()),
		),
		1, //limit to 1
		factory.SortAsc("rentals.rentalPeriod.startDate"),
	), &cars)
	if err != nil {
		return nil, err
	}

	if len(cars) == 0 {
		return nil, nil
	}
	rental := mappers.MapCarsFromDbToRentals(&cars, c.timeProvider)

	return &rental[0], nil
}

func (c *crud) GetTrunkAccess(ctx context.Context, vin model.Vin, token model.TrunkAccessToken) (*model.TrunkAccess,
	error) {

	var cars []entities.Car

	factory := c.db.GetFactory()

	err := c.db.Aggregate(
		ctx, c.collection, factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterAnd(
				factory.FilterEqual("rentals.trunkToken.token", token),
				factory.FilterEqual("_id", vin),
			),
			1,
			nil,
		),
		&cars,
	)

	if err != nil {
		return nil, err
	}

	if len(cars) == 0 {
		return nil, rentalErrors.ErrTrunkAccessDenied
	}

	rentals := mappers.MapCarFromDbToRentals(&cars[0], c.timeProvider)
	return rentals[0].Token, nil
}
