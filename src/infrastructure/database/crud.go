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

// ICRUD is a high level database interface. It directly maps to the business logic and abstracts away the
// database entities and the database connection.
type ICRUD interface {
	GetUnavailableCars(ctx context.Context, timePeriod model.TimePeriod) (*[]model.Vin, error)
	CreateRental(ctx context.Context, vin model.Vin, customerId model.CustomerId, timePeriod model.TimePeriod) error
	GetRentalsOfCustomer(ctx context.Context, customerID model.CustomerId) (*[]model.Rental, error)
	GetRental(ctx context.Context, rentalId model.RentalId) (*model.Rental, error)
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
