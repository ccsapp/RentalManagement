package main

import (
	"RentalManagement/infrastructure/database"
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/logic/model"
	"RentalManagement/testdata"
	"RentalManagement/testhelpers"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"
)

type ApiTestSuite struct {
	suite.Suite
	dbConnection       db.IConnection
	collection         string
	app                *echo.Echo
	config             *Config
	recordingFormatter *testhelpers.RecordingFormatter
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
	if err != nil {
		suite.T().Fatal(err.Error())
	}

	suite.app, err = newApp(suite.config, suite.dbConnection, dbConfig)
	if err != nil {
		suite.T().Fatal(err.Error())
	}
}

func (suite *ApiTestSuite) SetupTest() {
	suite.recordingFormatter = testhelpers.NewRecordingFormatter()
}

func (suite *ApiTestSuite) TearDownSuite() {
	// close the database connection when the program exits
	if err := suite.dbConnection.CleanUpDatabase(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *ApiTestSuite) TearDownTest() {
	// generate the sequence diagram for the test
	suite.recordingFormatter.SetOutFileName(suite.T().Name())
	suite.recordingFormatter.SetTitle(suite.T().Name())

	diagramFormatter := apitest.SequenceDiagram()
	diagramFormatter.Format(suite.recordingFormatter.GetRecorder())

	// clear the collection after each test
	if err := suite.dbConnection.DropCollection(context.Background(), suite.collection); err != nil {
		suite.T().Fatal(err)
	}
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (suite *ApiTestSuite) newApiTestWithMocks(mocks []*apitest.Mock) *apitest.APITest {
	return apitest.New().
		Mocks(mocks...).
		Debug().
		Handler(suite.app).
		Report(suite.recordingFormatter)
}

func (suite *ApiTestSuite) newApiTestWithCarMock() *apitest.APITest {
	return suite.newApiTestWithMocks(suite.newCarMock())
}

func (suite *ApiTestSuite) newCarMock() []*apitest.Mock {
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

func createRentalForCustomer(suite *ApiTestSuite, vin string, body string, customerId string) {
	suite.newApiTestWithCarMock().
		Post("/cars/"+vin+"/rentals").
		Query("customerId", customerId).
		JSON(body).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Body("").
		End()
}

func createRental(suite *ApiTestSuite, vin string, body string) {
	createRentalForCustomer(suite, vin, body, "d9ChwOvI")
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

	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingPast() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)

	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingFuture() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2123)

	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_carNotFound() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.UnknownVin+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_semanticallyInvalidTimePeriod() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSemanticInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_syntacticallyInvalidTimePeriod() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSyntaxInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidCustomerId() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "invalid").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidVin() {
	suite.newApiTestWithCarMock().
		Post("/cars/invalid/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_past() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod1900).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_beginPast() {
	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodLong).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_noRentals() {
	suite.newApiTestWithCarMock().
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
	suite.newApiTestWithCarMock().
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
	suite.newApiTestWithCarMock().
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
	suite.newApiTestWithCarMock().
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
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.EmptyArray).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_endDateBeforeStartDate() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("startDate", "2150-12-31T23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingStartDate() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingEndDate() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingBothDates() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidStartDate() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("startDate", "2150-12-3sdf23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidEndDate() {
	suite.newApiTestWithCarMock().
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2122-01-01s0blabla:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetCar_success() {
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar).
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.ExampleCarStaticResponse).
		End()
}

func (suite *ApiTestSuite) TestGetCar_unknownCar() {
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.UnknownVin).
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func (suite *ApiTestSuite) TestGetCar_invalidVin() {
	suite.newApiTestWithCarMock().
		Get("/cars/invalid").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetRentalOverview_success_noRentals() {
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.EmptyArray).
		End()
}

func (suite *ApiTestSuite) TestGetRentalOverview_success_oneRental() {
	createRentalForCustomer(suite, testdata.VinCar, testdata.TimePeriod2122, "d9ChwOvI")
	createRentalForCustomer(suite, testdata.VinCar, testdata.TimePeriod2150, "aDfd3Dae")
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRentalArray([]model.Rental{
			{
				Active: false,
				Car: &model.Car{
					Brand: "Audi",
					Model: "A3",
					Vin:   "WVWAA71K08W201030",
				},
				RentalPeriod: model.TimePeriod{
					EndDate:   time.Date(2123, 01, 01, 0, 0, 0, 0, time.UTC),
					StartDate: time.Date(2122, 01, 01, 0, 0, 0, 0, time.UTC),
				},
			}}, suite.T())).
		End()
}

func (suite *ApiTestSuite) TestGetRentalOverview_success_twoRentals() {
	createRentalForCustomer(suite, testdata.VinCar, testdata.TimePeriod2122, "d9ChwOvI")
	createRentalForCustomer(suite, testdata.VinCar, testdata.TimePeriod2150, "aDfd3Dae")
	createRentalForCustomer(suite, testdata.VinCar2, testdata.TimePeriod2150, "d9ChwOvI")
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRentalArray([]model.Rental{
			{
				Active: false,
				Car: &model.Car{
					Brand: "Audi",
					Model: "A3",
					Vin:   "WVWAA71K08W201030",
				},
				RentalPeriod: model.TimePeriod{
					EndDate:   time.Date(2123, 01, 01, 0, 0, 0, 0, time.UTC),
					StartDate: time.Date(2122, 01, 01, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				Active: false,
				Car: &model.Car{
					Brand: "Mercedes",
					Model: "B4",
					Vin:   "1FVNY5Y90HP312888",
				},
				RentalPeriod: model.TimePeriod{
					EndDate:   time.Date(2151, 01, 01, 0, 0, 0, 0, time.UTC),
					StartDate: time.Date(2150, 01, 01, 0, 0, 0, 0, time.UTC),
				},
			}}, suite.T())).
		End()
}

func returnsRentalArray(expectedRentals []model.Rental, t *testing.T) func(res *http.Response, _ *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var rentals []model.Rental
		if err := json.Unmarshal(bodyBytes, &rentals); err != nil {
			return err
		}

		// sort by Vin, descending
		sort.Slice(rentals, func(i, j int) bool {
			return rentals[i].Car.Vin > rentals[j].Car.Vin
		})

		for i := range rentals {
			rentals[i].Id = ""
		}

		assert.Equal(t, expectedRentals, rentals)

		return nil
	}
}

func (suite *ApiTestSuite) TestGetRentalOverview_invalidCustomerId() {
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "waytoolongcustomerid").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetRentalOverview_missingCustomerId() {
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}
