package main

import (
	"RentalManagement/infrastructure/database"
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/testdata"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

type ApiTestSuite struct {
	suite.Suite
	dbConnection db.IConnection
	collection   string
	app          *echo.Echo
	config       *Config
}

func (suite *ApiTestSuite) SetupSuite() {
	// load the environment variables for the database layer
	dbConfig, err := db.LoadConfigFromFile("testdata/testdb.env")
	if err != nil {
		suite.T().Fatal(err.Error())
	}

	suite.config = &Config{allowOrigins: []string{"*"}, domainServer: "https://carservice.kit.edu", domainTimeout: 1}

	// generate a collection name so that concurrent executions do not interfere
	dbConfig.CollectionPrefix = fmt.Sprintf("test-%d-", time.Now().Unix())
	suite.collection = dbConfig.CollectionPrefix + database.CollectionBaseName

	suite.dbConnection, err = db.NewDbConnection(dbConfig)

	suite.app, err = newApp(suite.config, suite.dbConnection, dbConfig)
	if err != nil {
		suite.T().Fatal(err.Error())
	}
}

func (suite *ApiTestSuite) TearDownSuite() {
	// close the database connection when the program exits
	if err := suite.dbConnection.CleanUpDatabase(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *ApiTestSuite) TearDownTest() {
	// clear the collection after each test
	if err := suite.dbConnection.DropCollection(context.Background(), suite.collection); err != nil {
		suite.T().Fatal(err)
	}
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func newApiTestWithMocks(handler http.Handler, name string, mocks []*apitest.Mock) *apitest.APITest {
	return apitest.New(name).
		Mocks(mocks...).
		Debug().
		Handler(handler).
		Report(apitest.SequenceDiagram())
}

func newCarMock(suite *ApiTestSuite) []*apitest.Mock {
	return []*apitest.Mock{
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars/" + testdata.VinCar).
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCar).End(),
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars/" + testdata.VinCar2).
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCar2).End(),
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars/" + testdata.UnknownVin).
			RespondWith().Status(http.StatusNotFound).End(),
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars").
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCarVins).End(),
	}
}

func createRental(suite *ApiTestSuite, vin string, body string) {
	newApiTestWithMocks(suite.app, "Create Rental", newCarMock(suite)).
		Post("/cars/"+vin+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(body).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Body("").
		End()
}

func (suite *ApiTestSuite) TestCreateRental_success_noRentals() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
}

func (suite *ApiTestSuite) TestCreateRental_success_nonConflictingRentalsExist() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2150)
	createRental(suite, testdata.VinCar, testdata.TimePeriod2150)
}

func (suite *ApiTestSuite) TestCreateRental_success_closeTimePeriod() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar, testdata.TimePeriod2123)
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)

	newApiTestWithMocks(suite.app, "Create conflicting Rental", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingPast() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)

	newApiTestWithMocks(suite.app, "Create conflicting Rental", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingFuture() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2123)

	newApiTestWithMocks(suite.app, "Create conflicting Rental", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_carNotFound() {
	newApiTestWithMocks(suite.app, "Create Rental with unknown car", newCarMock(suite)).
		Post("/cars/"+testdata.UnknownVin+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_semanticallyInvalidTimePeriod() {
	newApiTestWithMocks(suite.app, "Create Rental with invalid time period", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSemanticInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_syntacticallyInvalidTimePeriod() {
	newApiTestWithMocks(suite.app, "Create Rental with invalid time period", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSyntaxInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidCustomerId() {
	newApiTestWithMocks(suite.app, "Create Rental with invalid customer id", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "invalid").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidVin() {
	newApiTestWithMocks(suite.app, "Create Rental with invalid vin", newCarMock(suite)).
		Post("/cars/invalid/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_past() {
	newApiTestWithMocks(suite.app, "Create Rental in past", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod1900).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_beginPast() {
	newApiTestWithMocks(suite.app, "Create Rental with begin in past", newCarMock(suite)).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodLong).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_noRentals() {
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2123-01-21T17:32:28Z").
		Query("endDate", "2123-07-21T17:32:28Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.CarsAvailableBoth).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_bothAvailable() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar, testdata.TimePeriod2150)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2150)
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2123-01-01T00:00:00Z").
		Query("endDate", "2150-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.CarsAvailableBoth).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_secondBlockedInPast() {
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar, testdata.TimePeriod2150)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2150)
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2122-12-31T00:00:00Z").
		Query("endDate", "2149-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.CarsAvailableFirst).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_secondBlockedInFuture() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2150)
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2123-01-01T00:00:00Z").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.CarsAvailableFirst).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_bothBlocked() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2122)
	createRental(suite, testdata.VinCar, testdata.TimePeriod2150)
	createRental(suite, testdata.VinCar2, testdata.TimePeriod2150)
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.EmptyArray).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_endDateBeforeStartDate() {
	newApiTestWithMocks(suite.app, "Get Available Cars (invalid time period)", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2150-12-31T23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingStartDate() {
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingEndDate() {
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingBothDates() {
	newApiTestWithMocks(suite.app, "Get Available Cars", newCarMock(suite)).
		Get("/cars").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidStartDate() {
	newApiTestWithMocks(suite.app, "Get Available Cars (invalid time period)", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2150-12-3sdf23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidEndDate() {
	newApiTestWithMocks(suite.app, "Get Available Cars (invalid time period)", newCarMock(suite)).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2122-01-01s0blabla:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}
