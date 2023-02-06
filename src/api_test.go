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
	"hash"
	"hash/fnv"
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
	formatter    *recordingFormatter
}

type recordingFormatter struct {
	recorder *apitest.Recorder
	hash     hash.Hash32
}

func newRecordingFormatter() *recordingFormatter {
	return &recordingFormatter{recorder: apitest.NewTestRecorder(), hash: fnv.New32a()}
}

func (rf *recordingFormatter) Format(recorder *apitest.Recorder) {
	// append the events to the existing events
	rf.recorder.Events = append(rf.recorder.Events, recorder.Events...)
	_, err := rf.hash.Write(([]byte)(recorder.Meta["hash"].(string)))
	if err != nil {
		// this cannot happen
		panic(err)
	}

	meta := make(map[string]interface{})
	meta["hash"] = fmt.Sprintf("%d", rf.hash.Sum32())

	rf.recorder.AddMeta(meta)
}

func (rf *recordingFormatter) SetTitle(title string) {
	rf.recorder.AddTitle(title)
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
	suite.formatter = newRecordingFormatter()
}

func (suite *ApiTestSuite) TearDownSuite() {
	// close the database connection when the program exits
	if err := suite.dbConnection.CleanUpDatabase(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *ApiTestSuite) TearDownTest() {
	// generate the sequence diagram for the test
	suite.formatter.SetTitle(suite.T().Name())
	diagramFormatter := apitest.SequenceDiagram()
	diagramFormatter.Format(suite.formatter.recorder)

	// clear the collection after each test
	if err := suite.dbConnection.DropCollection(context.Background(), suite.collection); err != nil {
		suite.T().Fatal(err)
	}
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (suite *ApiTestSuite) newApiTestWithMocks(handler http.Handler, mocks []*apitest.Mock) *apitest.APITest {
	return apitest.New().
		Mocks(mocks...).
		Debug().
		Handler(handler).
		Report(suite.formatter)
}

func (suite *ApiTestSuite) newApiTestWithCarMock(handler http.Handler) *apitest.APITest {
	return suite.newApiTestWithMocks(handler, suite.newCarMock())
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

func createRental(suite *ApiTestSuite, vin string, body string) {
	suite.newApiTestWithCarMock(suite.app).
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

	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingPast() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2122)

	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingFuture() {
	createRental(suite, testdata.VinCar, testdata.TimePeriod2123)

	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_carNotFound() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.UnknownVin+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_semanticallyInvalidTimePeriod() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSemanticInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_syntacticallyInvalidTimePeriod() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodSyntaxInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidCustomerId() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "invalid").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_invalidVin() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/invalid/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_past() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod1900).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_beginPast() {
	suite.newApiTestWithCarMock(suite.app).
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriodLong).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_success_noRentals() {
	suite.newApiTestWithCarMock(suite.app).
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
	suite.newApiTestWithCarMock(suite.app).
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
	suite.newApiTestWithCarMock(suite.app).
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
	suite.newApiTestWithCarMock(suite.app).
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
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.EmptyArray).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_endDateBeforeStartDate() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("startDate", "2150-12-31T23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingStartDate() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("endDate", "2150-12-31T23:59:59Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingEndDate() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_missingBothDates() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidStartDate() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("startDate", "2150-12-3sdf23:59:59Z").
		Query("endDate", "2122-01-01T00:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetAvailableCars_invalidEndDate() {
	suite.newApiTestWithCarMock(suite.app).
		Get("/cars").
		Query("startDate", "2122-01-01T00:00:00Z").
		Query("endDate", "2122-01-01s0blabla:00:00Z").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}
