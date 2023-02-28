package database

import (
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/infrastructure/database/entities"
	"RentalManagement/infrastructure/database/mappers"
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

var collectionPrefix = "collectionPrefix"
var dbConfig = &db.Config{CollectionPrefix: collectionPrefix}

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

	crud := NewICRUD(mockConnection, dbConfig, mockTime)

	err := crud.CreateRental(ctx, vin, customerId, timePeriod2023)
	assert.ErrorIs(t, err, rentalErrors.ErrConflictingRentalExists)
}

func TestCrud_GetRentalsOfCustomer_success_NoRentals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	customerId := "jJ8mNg6Z"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	var car1 = entities.Car{
		Vin:     "WVWAA71K08W201030",
		Rentals: []entities.Rental{},
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.customer", customerId),
			-1,
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{car1}).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRentals, err := crud.GetRentalsOfCustomer(ctx, customerId)

	assert.Nil(t, err)
	assert.Equal(t, &[]model.Rental{}, returnedRentals)
}

func TestCrud_GetRentalsOfCustomer_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConnection := mocks.NewMockIConnection(ctrl)

	factory := db.PseudoFactory{}

	customerId := "jJ8mNg6Z"

	var car1 = entities.Car{
		Vin: "WVWAA71K08W201030",
		Rentals: []entities.Rental{
			{
				RentalId:   "rZ6IIwcD",
				CustomerId: customerId,
				RentalPeriod: entities.TimePeriod{
					EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
					StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
				},
				TrunkToken: nil,
			},
		},
	}

	var car2 = entities.Car{
		Vin: "AVWAA71K08W201031",
		Rentals: []entities.Rental{
			{
				RentalId:   "rZ6I8waD",
				CustomerId: customerId,
				RentalPeriod: entities.TimePeriod{
					EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
					StartDate: time.Date(2023, 1, 2, 3, 0, 0, 0, time.UTC),
				},
				TrunkToken: &entities.TrunkAccessToken{
					Token: "bumrLuCMbumrLuCMbumrLuCM",
					ValidityPeriod: entities.TimePeriod{
						EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
						StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	var rentalModel1 = model.Rental{
		State:    model.ACTIVE,
		Car:      &model.Car{Vin: "WVWAA71K08W201030"},
		Id:       "rZ6IIwcD",
		Customer: &model.Customer{CustomerId: customerId},
		RentalPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
		Token: nil,
	}

	var rentalModel2 = model.Rental{
		State:    model.EXPIRED,
		Car:      &model.Car{Vin: "AVWAA71K08W201031"},
		Id:       "rZ6I8waD",
		Customer: &model.Customer{CustomerId: customerId},
		RentalPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 3, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 1, 2, 3, 0, 0, 0, time.UTC),
		},
		Token: &model.TrunkAccess{
			Token: "bumrLuCMbumrLuCMbumrLuCM",
			ValidityPeriod: model.TimePeriod{
				EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
				StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
			},
		},
	}

	var cars = []entities.Car{car1, car2}
	var rentals = []model.Rental{rentalModel1, rentalModel2}
	var currentDate = time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(currentDate)
	mockTimeProvider.EXPECT().Now().Return(currentDate)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.customer", customerId),
			-1,
			nil,
		),
		gomock.Any(),
	).SetArg(3, cars).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRentals, err := crud.GetRentalsOfCustomer(ctx, customerId)

	assert.Nil(t, err)
	assert.Equal(t, &rentals, returnedRentals)
}

func TestCrud_GetRentalsOfCustomer_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	dbError := errors.New("db error")
	customerId := "jJ8mNg6Z"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRentals, err := crud.GetRentalsOfCustomer(ctx, customerId)

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, returnedRentals)
}

func TestCrud_SetTrunkToken_success_noRestriction_existingToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2023, 6, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: &entities.TrunkAccessToken{
			Token: "bumrLuCMbumrLuCMbumrLuCM",
			ValidityPeriod: entities.TimePeriod{
				StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.Nil(t, err)
	assert.Equal(t, &newToken, retToken)
}

func TestCrud_SetTrunkToken_success_noRestriction_newToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2023, 6, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.Nil(t, err)
	assert.Equal(t, &newToken, retToken)
}

func TestCrud_SetTrunkToken_success_restriction_newToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2023, 6, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenRestricted := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newTokenRestricted)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.Nil(t, err)
	assert.Equal(t, &newTokenRestricted, retToken)
}

func TestCrud_SetTrunkToken_optimisticLockingError_recoverAfter1_restrict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	currentTime := time.Date(2023, 4, 1, 1, 0, 0, 0, time.UTC)
	mockTimeProvider.EXPECT().Now().Return(currentTime).Times(2)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().GetFactory().Return(&factory)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(db.NoDocumentsError)

	existingRentalAfterError := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "remotsuc",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	// CRUD should retry by first retrieving the rental again
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRentalAfterError},
		},
	}).Return(nil)

	// the token should be restricted to the new rental period after the optimistic locking error
	newTokenRestricted := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity = *mappers.MapTokenToDb(&newTokenRestricted)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRentalAfterError),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.Nil(t, err)
	assert.Equal(t, &newTokenRestricted, retToken)
}

func TestCrud_SetTrunkToken_optimisticLockingError_recoverAfter1_rentalDisappears(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	currentTime := time.Date(2023, 4, 1, 1, 0, 0, 0, time.UTC)
	mockTimeProvider.EXPECT().Now().Return(currentTime).Times(1)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().GetFactory().Return(&factory)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(db.NoDocumentsError)

	// CRUD should retry by first retrieving the rental again
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotFound)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_optimisticLockingError_recoverAfter1_rentalExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	currentTime := time.Date(2023, 4, 1, 1, 0, 0, 0, time.UTC)
	mockTimeProvider.EXPECT().Now().Return(currentTime).Times(2)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().GetFactory().Return(&factory)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(db.NoDocumentsError)

	// the rental is now expired
	existingRentalAfterError := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "remotsuc",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	// CRUD should retry by first retrieving the rental again
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRentalAfterError},
		},
	}).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotActive)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_optimisticLockingError_recoverAfter2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	currentTime := time.Date(2023, 4, 1, 1, 0, 0, 0, time.UTC)
	mockTimeProvider.EXPECT().Now().Return(currentTime).Times(3)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory).Times(3)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil).Times(3)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(db.NoDocumentsError).Times(2)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.Nil(t, err)
	assert.Equal(t, &newToken, retToken)
}

func TestCrud_SetTrunkToken_optimisticLockingError_failAfter3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	currentTime := time.Date(2023, 4, 1, 1, 0, 0, 0, time.UTC)
	mockTimeProvider.EXPECT().Now().Return(currentTime).Times(3)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory).Times(3)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil).Times(3)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(db.NoDocumentsError).Times(3)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, OptimisticLockingError)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_rentalNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{}).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", model.TrunkAccess{})

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotFound)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_dbError_aggregate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)

	dbError := errors.New("db error")

	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", model.TrunkAccess{})

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_dbError_update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2023, 6, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	newTokenEntity := *mappers.MapTokenToDb(&newToken)

	dbError := errors.New("db error")

	mockConnection.EXPECT().UpdateOne(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.FilterElementMatch(
			"rentals",
			factory.FilterMatch(existingRental),
		),
		factory.UpdateMatchingArrayElement(
			"rentals",
			"trunkToken",
			newTokenEntity,
		),
		false, // no upsert
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_periodNotOverlapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2023, 6, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotOverlapping)
	assert.Nil(t, retToken)
}

func TestCrud_SetTrunkToken_rentalNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	mockConnection := mocks.NewMockIConnection(ctrl)

	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(
		time.Date(2100, 1, 1, 1, 0, 0, 0, time.UTC),
	)

	existingRental := entities.Rental{
		RentalId:   "rentalId",
		CustomerId: "customer",
		RentalPeriod: entities.TimePeriod{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		TrunkToken: nil,
	}

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", "rentalId"),
			1, // limit to 1
			nil,
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{
		{
			Vin:     "AVWAA71K08W201031",
			Rentals: []entities.Rental{existingRental},
		},
	}).Return(nil)

	newToken := model.TrunkAccess{
		Token: "thisIsTheNewToken1234567",
		ValidityPeriod: model.TimePeriod{
			StartDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	retToken, err := crud.SetTrunkToken(ctx, "rentalId", newToken)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotActive)
	assert.Nil(t, retToken)
}

func TestCrud_GetRental_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConnection := mocks.NewMockIConnection(ctrl)

	factory := db.PseudoFactory{}

	rentalId := "rZ6IIwcD"

	var car = entities.Car{
		Vin: "WVWAA71K08W201030",
		Rentals: []entities.Rental{
			{
				RentalId:   rentalId,
				CustomerId: "jJ8mNg6Z",
				RentalPeriod: entities.TimePeriod{
					EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
					StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
				},
				TrunkToken: &entities.TrunkAccessToken{
					Token: "bumrLuCMbumrLuCMbumrLuCM",
					ValidityPeriod: entities.TimePeriod{
						EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
						StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	var rental = model.Rental{
		State:    model.ACTIVE,
		Car:      &model.Car{Vin: "WVWAA71K08W201030"},
		Id:       rentalId,
		Customer: &model.Customer{CustomerId: "jJ8mNg6Z"},
		RentalPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
		Token: &model.TrunkAccess{
			Token: "bumrLuCMbumrLuCMbumrLuCM",
			ValidityPeriod: model.TimePeriod{
				EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
				StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
			},
		},
	}

	var cars = []entities.Car{car}
	var currentDate = time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(currentDate)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterEqual("rentals.rentalId", rentalId),
			1,
			nil,
		),
		gomock.Any(),
	).SetArg(3, cars).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetRental(ctx, rentalId)

	assert.Nil(t, err)
	assert.Equal(t, &rental, returnedRental)
}

func TestCrud_GetRental_rentalIdNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	rentalId := "norental"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
	).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetRental(ctx, rentalId)

	assert.ErrorIs(t, err, rentalErrors.ErrRentalNotFound)
	assert.Nil(t, returnedRental)
}

func TestCrud_GetRental_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	dbError := errors.New("db error")
	rentalId := "rZ6IIwcD"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetRental(ctx, rentalId)

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, returnedRental)
}

func TestCrud_GetNextRental_success_exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	exampleCar := entities.Car{
		Vin: "WVWAA71K08W201030",
		Rentals: []entities.Rental{
			{
				RentalId:   "rZ6IIwcD",
				CustomerId: "jJ8mNg6Z",
				RentalPeriod: entities.TimePeriod{
					EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
					StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
				},
				TrunkToken: nil,
			},
		},
	}
	var cars = []entities.Car{exampleCar}
	exampleRental := model.Rental{
		State:    model.ACTIVE,
		Car:      &model.Car{Vin: "WVWAA71K08W201030"},
		Id:       "rZ6IIwcD",
		Customer: &model.Customer{CustomerId: "jJ8mNg6Z"},
		RentalPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
		},
		Token: nil,
	}
	currentDate := time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockTimeProvider.EXPECT().Now().Return(currentDate)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterAnd(
				factory.FilterEqual("_id", "WVWAA71K08W201030"),
				factory.FilterGreater("rentals.rentalPeriod.endDate", currentDate),
			),
			1,
			factory.SortAsc("rentals.rentalPeriod.startDate"),
		),
		gomock.Any(),
	).SetArg(3, cars).Return(nil)
	mockTimeProvider.EXPECT().Now().Return(currentDate)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetNextRental(ctx, "WVWAA71K08W201030")

	assert.Nil(t, err)
	assert.Equal(t, &exampleRental, returnedRental)
}

func TestCrud_GetNextRental_success_notExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}

	currentDate := time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockTimeProvider.EXPECT().Now().Return(currentDate)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterAnd(
				factory.FilterEqual("_id", "WVWAA71K08W201030"),
				factory.FilterGreater("rentals.rentalPeriod.endDate", currentDate),
			),
			1,
			factory.SortAsc("rentals.rentalPeriod.startDate"),
		),
		gomock.Any(),
	).SetArg(3, []entities.Car{}).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetNextRental(ctx, "WVWAA71K08W201030")

	assert.Nil(t, err)
	assert.Nil(t, returnedRental)
}

func TestCrud_GetNextRental_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	dbError := errors.New("db error")

	currentDate := time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockTimeProvider.EXPECT().Now().Return(currentDate)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterAnd(
				factory.FilterEqual("_id", "WVWAA71K08W201030"),
				factory.FilterGreater("rentals.rentalPeriod.endDate", currentDate),
			),
			1,
			factory.SortAsc("rentals.rentalPeriod.startDate"),
		),
		gomock.Any(),
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedRental, err := crud.GetNextRental(ctx, "WVWAA71K08W201030")

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, returnedRental)
}

func TestCrud_GetTrunkAccess_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockConnection := mocks.NewMockIConnection(ctrl)

	factory := db.PseudoFactory{}

	token := "bumrLuCMbumrLuCMbumrLuCM"
	vin := "WVWAA71K08W201030"

	var car = entities.Car{
		Vin: vin,
		Rentals: []entities.Rental{
			{
				RentalId:   "rZ6IIwcD",
				CustomerId: "jJ8mNg6Z",
				RentalPeriod: entities.TimePeriod{
					EndDate:   time.Date(2023, 4, 3, 1, 0, 0, 0, time.UTC),
					StartDate: time.Date(2023, 3, 2, 3, 0, 0, 0, time.UTC),
				},
				TrunkToken: &entities.TrunkAccessToken{
					Token: token,
					ValidityPeriod: entities.TimePeriod{
						EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
						StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	var access = model.TrunkAccess{
		Token: "bumrLuCMbumrLuCMbumrLuCM",
		ValidityPeriod: model.TimePeriod{
			EndDate:   time.Date(2023, 2, 3, 1, 0, 0, 0, time.UTC),
			StartDate: time.Date(2023, 2, 2, 3, 0, 0, 0, time.UTC),
		},
	}

	var cars = []entities.Car{car}

	var currentDate = time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)
	mockTimeProvider.EXPECT().Now().Return(currentDate)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		factory.ArrayFilterAggregation(
			"rentals",
			factory.FilterAnd(
				factory.FilterEqual("rentals.trunkToken.token", token),
				factory.FilterEqual("_id", vin),
			),
			1,
			nil,
		),
		gomock.Any(),
	).SetArg(3, cars).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedAccess, err := crud.GetTrunkAccess(ctx, vin, token)

	assert.Nil(t, err)
	assert.Equal(t, &access, returnedAccess)
}

func TestCrud_GetTrunkAccess_trunkAccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	token := "not token"
	vin := "WVWAA71K08W201030"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
	).SetArg(3, []entities.Car{}).Return(nil)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedAccess, err := crud.GetTrunkAccess(ctx, vin, token)

	assert.ErrorIs(t, err, rentalErrors.ErrTrunkAccessDenied)
	assert.Nil(t, returnedAccess)
}

func TestCrud_GetTrunkAccess_databaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	factory := db.PseudoFactory{}
	dbError := errors.New("db error")
	token := "bumrLuCMbumrLuCMbumrLuCM"
	vin := "WVWAA71K08W201030"

	mockConnection := mocks.NewMockIConnection(ctrl)
	mockTimeProvider := mocks.NewMockITimeProvider(ctrl)

	mockConnection.EXPECT().GetFactory().Return(&factory)
	mockConnection.EXPECT().Aggregate(
		ctx,
		collectionPrefix+CollectionBaseName,
		gomock.Any(),
		gomock.Any(),
	).Return(dbError)

	crud := NewICRUD(mockConnection, dbConfig, mockTimeProvider)
	returnedAccess, err := crud.GetTrunkAccess(ctx, vin, token)

	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, returnedAccess)
}
