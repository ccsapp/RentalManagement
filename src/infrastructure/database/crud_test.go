package database

import (
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/infrastructure/database/entities"
	"RentalManagement/logic/model"
	"RentalManagement/logic/rentalErrors"
	"RentalManagement/mocks"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var timePeriod2023 = model.TimePeriod{
	StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
}

func TestCrud_GetUnavailableCars_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"
	timePeriod := model.TimePeriod{
		StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2052, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	var cars = []entities.Car{
		{
			Vin: "SAJWA0ES6DPS56028",
		},
		{
			Vin: "1G1ZB5ST5GF123456",
		},
	}

	var expectedVins = []model.Vin{"SAJWA0ES6DPS56028", "1G1ZB5ST5GF123456"}

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().FindMany(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterAnd(
				factory.FilterLess("rentalPeriod.startDate", timePeriod.EndDate),
				factory.FilterGreater("rentalPeriod.endDate", timePeriod.StartDate),
			),
		),
		&db.Options{Projection: factory.ProjectionID()},
		gomock.Any(),
	).SetArg(4, cars).Return(nil)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)
	vins, err := crud.GetUnavailableCars(ctx, timePeriod)

	assert.Nil(t, err)
	assert.Equal(t, expectedVins, *vins)
}

func TestCrud_GetUnavailableCars_successEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"
	timePeriod := model.TimePeriod{
		StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2052, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().FindMany(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterAnd(
				factory.FilterLess("rentalPeriod.startDate", timePeriod.EndDate),
				factory.FilterGreater("rentalPeriod.endDate", timePeriod.StartDate),
			),
		),
		&db.Options{Projection: factory.ProjectionID()},
		gomock.Any(),
	).SetArg(4, []entities.Car{}).Return(nil)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)
	vins, err := crud.GetUnavailableCars(ctx, timePeriod)

	assert.Nil(t, err)

	// nil means empty
	assert.Equal(t, []string{}, *vins)
}

func TestCrud_GetUnavailableCars_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	databaseError := errors.New("database error")

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().FindMany(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterAnd(
				factory.FilterLess("rentalPeriod.startDate", timePeriod2023.EndDate),
				factory.FilterGreater("rentalPeriod.endDate", timePeriod2023.StartDate),
			),
		),
		&db.Options{Projection: factory.ProjectionID()},
		gomock.Any(),
	).Return(databaseError)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)
	vins, err := crud.GetUnavailableCars(ctx, timePeriod2023)

	assert.ErrorIs(t, err, databaseError)
	assert.Nil(t, vins)
}

func TestCrud_CreateRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"

	vin := "SAJWA0ES6DPS56028"
	customerId := "jJ8mNg6Z"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterAnd(
			factory.FilterEqual("_id", vin),
			factory.FilterNot(
				factory.FilterElementMatch(
					"rentals",
					factory.FilterAnd(
						factory.FilterLess("rentalPeriod.startDate", timePeriod2023.EndDate),
						factory.FilterGreater("rentalPeriod.endDate", timePeriod2023.StartDate),
					),
				),
			),
		),
		gomock.Any(),
		true,
	).Do(func(_ context.Context, _ string, _ interface{}, update db.Update, _ bool) {
		// since the rental id is random, we can't check it
		// and need to check everything else manually
		fieldName, value, err := factory.UnpackPushUpdate(update)
		assert.Nil(t, err)
		assert.Equal(t, "rentals", fieldName)

		rental, ok := value.(entities.Rental)
		assert.True(t, ok)
		assert.Equal(t, customerId, rental.CustomerId)
		assert.Equal(t, timePeriod2023.StartDate, rental.RentalPeriod.StartDate)
		assert.Equal(t, timePeriod2023.EndDate, rental.RentalPeriod.EndDate)
		assert.Equal(t, 8, len(rental.RentalId))
	}).Return(nil)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)

	err := crud.CreateRental(ctx, vin, customerId, timePeriod2023)
	assert.Nil(t, err)
}

func TestCrud_CreateRental_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"

	vin := "SAJWA0ES6DPS56028"
	customerId := "jJ8mNg6Z"

	databaseError := errors.New("database error")

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
		true,
	).Return(databaseError)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)

	err := crud.CreateRental(ctx, vin, customerId, timePeriod2023)
	assert.ErrorIs(t, err, databaseError)
}

func TestCrud_CreateRental_conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	factory := &db.PseudoFactory{}

	mockTime := mocks.NewMockITimeProvider(ctrl)

	collectionPrefix := "collectionPrefix"

	vin := "SAJWA0ES6DPS56028"
	customerId := "jJ8mNg6Z"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockConnection.EXPECT().GetFactory().Return(factory)
	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
		true,
	).Return(db.DuplicateKeyError)

	dbConfig := &db.Config{CollectionPrefix: collectionPrefix}
	crud := NewICRUD(mockConnection, dbConfig, mockTime)

	err := crud.CreateRental(ctx, vin, customerId, timePeriod2023)
	assert.ErrorIs(t, err, rentalErrors.ErrConflictingRentalExists)
}
