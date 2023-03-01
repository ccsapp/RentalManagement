package main

import (
	"RentalManagement/environment"
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
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

const carServerUrl = "https://carservice.kit.edu"

type ApiTestSuite struct {
	suite.Suite
	dbConnection       db.IConnection
	collection         string
	app                *echo.Echo
	recordingFormatter *testhelpers.RecordingFormatter
}

func (suite *ApiTestSuite) SetupSuite() {
	environment.SetupTestingEnvironment(carServerUrl)

	// generate a collection name so that concurrent executions do not interfere
	collectionPrefix := fmt.Sprintf("test-%d-", time.Now().Unix())
	environment.GetEnvironment().SetAppCollectionPrefix(collectionPrefix)
	suite.collection = collectionPrefix + database.CollectionBaseName

	var err error
	suite.dbConnection, err = db.NewDbConnection(environment.GetEnvironment())
	if err != nil {
		suite.handleDbConnectionError(err)
	}

	suite.app, err = newApp(suite.dbConnection)
	if err != nil {
		suite.T().Fatal(err.Error())
	}
}

func (suite *ApiTestSuite) handleDbConnectionError(err error) {
	// if local setup mode is disabled, we fail without any additional checks
	if !environment.GetEnvironment().IsLocalSetupMode() {
		suite.T().Fatal(err.Error())
	}

	running, dockerErr := testhelpers.IsMongoDbContainerRunning()
	if dockerErr != nil {
		suite.T().Fatal(dockerErr.Error())
	}
	if !running {
		suite.T().Fatal("MongoDB container is not running. " +
			"Please start it with 'docker compose up -d' and try again.")
	}
	fmt.Println("MongoDB container seems to be running, but the connection could not be established. " +
		"Please check the logs for more information.")
	suite.T().Fatal(err.Error())
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
			Get(environment.GetEnvironment().GetCarServerUrl() + "/cars/" + testdata.VinCar).
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCar).End(),
		apitest.NewMock().
			Get(environment.GetEnvironment().GetCarServerUrl() + "/cars/" + testdata.VinCar2).
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCar2).End(),
		apitest.NewMock().
			Get(environment.GetEnvironment().GetCarServerUrl() + "/cars/" + testdata.UnknownVin).
			RespondWith().Status(http.StatusNotFound).End(),
		apitest.NewMock().
			Get(environment.GetEnvironment().GetCarServerUrl() + "/cars").
			RespondWith().Status(http.StatusOK).JSON(testdata.ExampleCarVins).End(),
		apitest.NewMock().
			Put(environment.GetEnvironment().GetCarServerUrl() + "/cars/" + testdata.VinCar + "/trunkLock").
			RespondWith().Status(http.StatusNoContent).End(),
	}
}

func (suite *ApiTestSuite) createRentalForCustomer(vin string, body string, customerId string) {
	suite.newApiTestWithCarMock().
		Post("/cars/"+vin+"/rentals").
		Query("customerId", customerId).
		JSON(body).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Body("").
		End()
}

func (suite *ApiTestSuite) createRental(vin string, body string) {
	suite.createRentalForCustomer(vin, body, "d9ChwOvI")
}

func (suite *ApiTestSuite) getRentalOverview(customerId model.CustomerId) []model.Rental {
	var rentals []model.Rental
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", customerId).
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(mapOverviewToRentals(&rentals)).
		End()
	return rentals
}

func mapOverviewToRentals(rentals *[]model.Rental) func(*http.Response, *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()
		return json.NewDecoder(res.Body).Decode(rentals)
	}
}

func (suite *ApiTestSuite) getRentalDetailed(id model.RentalId) model.Rental {
	var rental model.Rental
	suite.newApiTestWithCarMock().
		Get("/rentals/" + id).
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(mapDetailedToRental(&rental)).
		End()
	return rental
}

func mapDetailedToRental(rental *model.Rental) func(*http.Response, *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()
		return json.NewDecoder(res.Body).Decode(rental)
	}
}

func (suite *ApiTestSuite) grantTrunkAccess(rentalId model.RentalId, timePeriod model.TimePeriod) model.TrunkAccess {
	var trunkAccess model.TrunkAccess
	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(timePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		Assert(mapGrantedToTrunkAccess(&trunkAccess)).
		End()
	return trunkAccess
}

func mapGrantedToTrunkAccess(trunkAccess *model.TrunkAccess) func(*http.Response, *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()
		return json.NewDecoder(res.Body).Decode(trunkAccess)
	}
}

func (suite *ApiTestSuite) TestCreateRental_success_noRentals() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
}

func (suite *ApiTestSuite) TestCreateRental_success_nonConflictingRentalsExist() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2150)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)
}

func (suite *ApiTestSuite) TestCreateRental_success_closeTimePeriod() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2123)
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)

	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingPast() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)

	suite.newApiTestWithCarMock().
		Post("/cars/"+testdata.VinCar+"/rentals").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.TimePeriod2122To23).
		Expect(suite.T()).
		Status(http.StatusConflict).
		End()
}

func (suite *ApiTestSuite) TestCreateRental_conflictingRentalsExist_overlappingFuture() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2123)

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
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2150)
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
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2150)
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
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2150)
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
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)
	suite.createRental(testdata.VinCar2, testdata.TimePeriod2150)
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
	now := time.Now().UTC().Round(time.Millisecond)

	timePeriodFuture := model.TimePeriod{
		StartDate: now.Add(10 * time.Hour),
		EndDate:   now.Add(12 * time.Hour),
	}

	marshalledPeriodFromFuture, _ := json.Marshal(timePeriodFuture)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriodFromFuture), "d9ChwOvI")
	suite.createRentalForCustomer(testdata.VinCar, testdata.TimePeriod2150, "aDfd3Dae")
	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRentalArray([]model.Rental{
			{
				State: model.UPCOMING,
				Car: &model.Car{
					Brand: "Audi",
					Model: "A3",
					Vin:   "WVWAA71K08W201030",
				},
				RentalPeriod: timePeriodFuture,
			}}, suite.T())).
		End()
}

func (suite *ApiTestSuite) TestGetRentalOverview_success_twoRentals() {
	now := time.Now().UTC().Round(time.Millisecond)

	timePeriod1 := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Hour),
	}

	marshalledPeriod1, _ := json.Marshal(timePeriod1)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriod1), "d9ChwOvI")
	suite.createRentalForCustomer(testdata.VinCar, testdata.TimePeriod2150, "aDfd3Dae")

	now = time.Now().UTC().Round(time.Millisecond)

	timePeriod2 := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Hour),
	}

	marshalledPeriod2, _ := json.Marshal(timePeriod2)

	suite.createRentalForCustomer(testdata.VinCar2, string(marshalledPeriod2), "d9ChwOvI")

	time.Sleep(15*time.Millisecond - time.Since(now))

	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRentalArray([]model.Rental{
			{
				State: model.ACTIVE,
				Car: &model.Car{
					Brand: "Audi",
					Model: "A3",
					Vin:   "WVWAA71K08W201030",
				},
				RentalPeriod: timePeriod1,
			},
			{
				State: model.ACTIVE,
				Car: &model.Car{
					Brand: "Mercedes",
					Model: "B4",
					Vin:   "1FVNY5Y90HP312888",
				},
				RentalPeriod: timePeriod2,
			}}, suite.T())).
		End()
}

func returnsRentalArray(expectedRentals []model.Rental, t *testing.T) func(res *http.Response, _ *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()

		var rentals []model.Rental
		if err := json.NewDecoder(res.Body).Decode(&rentals); err != nil {
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

func (suite *ApiTestSuite) TestGetRentalOverview_expiredRental() {
	now := time.Now().UTC().Round(time.Millisecond)

	timePeriodExpired := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   now.Add(15 * time.Millisecond).UTC().Round(time.Millisecond),
	}

	marshalledPeriodExpired, _ := json.Marshal(timePeriodExpired)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriodExpired), "d9ChwOvI")

	time.Sleep(15 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Get("/rentals").
		Query("customerId", "d9ChwOvI").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRentalArray([]model.Rental{
			{
				State: model.EXPIRED,
				Car: &model.Car{
					Brand: "Audi",
					Model: "A3",
					Vin:   "WVWAA71K08W201030",
				},
				RentalPeriod: timePeriodExpired,
			}}, suite.T())).
		End()
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

func (suite *ApiTestSuite) TestGetRentalStatus_success_upcomingRental() {
	suite.createRentalForCustomer(testdata.VinCar, testdata.TimePeriod2122, "d9ChwOvI")

	rentalId := suite.getRentalOverview("d9ChwOvI")[0].Id

	suite.newApiTestWithCarMock().
		Get("/rentals/" + rentalId).
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(strings.Replace(
			testdata.CustomerRentalUpcoming, "WillBeReplacedDynamicallyDuringTesting_RentalID", rentalId, 1)).
		End()
}

func (suite *ApiTestSuite) TestGetRentalStatus_success_expiredRental() {
	periodFromNow := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Now().Add(15 * time.Millisecond).UTC().Round(time.Millisecond),
	}

	marshalledPeriodFromNow, _ := json.Marshal(periodFromNow)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriodFromNow), "d9ChwOvI")

	time.Sleep(15 * time.Millisecond)

	rentalId := suite.getRentalOverview("d9ChwOvI")[0].Id

	expectedRental := strings.Replace(
		testdata.CustomerRentalExpired, "\"WillBeReplacedDynamicallyDuringTesting_RentalPeriod\"",
		string(marshalledPeriodFromNow), 1)
	expectedRental = strings.Replace(
		expectedRental, "WillBeReplacedDynamicallyDuringTesting_RentalID", rentalId, 1)

	suite.newApiTestWithCarMock().
		Get("/rentals/" + rentalId).
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(expectedRental).
		End()
}

func (suite *ApiTestSuite) TestGetRentalStatus_success_activeRental() {
	periodFromNow := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledPeriodFromNow, _ := json.Marshal(periodFromNow)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriodFromNow), "d9ChwOvI")

	time.Sleep(15 * time.Millisecond)

	rentalId := suite.getRentalOverview("d9ChwOvI")[0].Id

	expectedRental := strings.Replace(
		testdata.CustomerRentalActive, "\"WillBeReplacedDynamicallyDuringTesting_RentalPeriod\"",
		string(marshalledPeriodFromNow), 1)
	expectedRental = strings.Replace(
		expectedRental, "WillBeReplacedDynamicallyDuringTesting_RentalID", rentalId, 1)

	suite.newApiTestWithCarMock().
		Get("/rentals/" + rentalId).
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(expectedRental).
		End()
}

func (suite *ApiTestSuite) TestGetRentalStatus_unknownRentalId() {
	suite.newApiTestWithCarMock().
		Get("/rentals/unkownid").
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func (suite *ApiTestSuite) TestGetRentalStatus_invalidRentalId() {
	suite.newApiTestWithCarMock().
		Get("/rentals/waytoolongrentalid").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) assertToken(period model.TimePeriod) func(*http.Response, *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()

		var token model.TrunkAccess
		if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
			return err
		}

		assert.Equal(suite.T(), period, token.ValidityPeriod)
		assert.Equal(suite.T(), 24, len(token.Token))

		return nil
	}
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_match() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	rentalBefore := suite.getRentalDetailed(rentalId)

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(timePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()

	rental := suite.getRentalDetailed(rentalId)
	assert.Equal(suite.T(), timePeriod, rental.Token.ValidityPeriod)
	assert.Equal(suite.T(), 24, len(rental.Token.Token))
	assert.Equal(suite.T(), rentalBefore.Car, rental.Car)
	assert.Equal(suite.T(), rentalBefore.State, rental.State)
	assert.Equal(suite.T(), rentalBefore.RentalPeriod, rental.RentalPeriod)
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_overwrite() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(timePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()

	timePeriod.StartDate = timePeriod.StartDate.Add(24 * time.Hour)
	timePeriod.EndDate = timePeriod.EndDate.Add(-24 * time.Hour)

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(timePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()

	rental := suite.getRentalDetailed(rentalId)
	assert.Equal(suite.T(), timePeriod, rental.Token.ValidityPeriod)
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_inside() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	timePeriod.StartDate = timePeriod.StartDate.Add(24 * time.Hour)
	timePeriod.EndDate = timePeriod.EndDate.Add(-24 * time.Hour)

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(timePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_outside() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	outsideTimePeriod := model.TimePeriod{
		StartDate: timePeriod.StartDate.Add(-24 * time.Hour),
		EndDate:   timePeriod.EndDate.Add(24 * time.Hour),
	}

	rentalId := suite.getRentalOverview("customer")[0].Id

	rentalBefore := suite.getRentalDetailed(rentalId)

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(outsideTimePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()

	rental := suite.getRentalDetailed(rentalId)
	assert.Equal(suite.T(), timePeriod, rental.Token.ValidityPeriod)
	assert.Equal(suite.T(), 24, len(rental.Token.Token))
	assert.Equal(suite.T(), rentalBefore.Car, rental.Car)
	assert.Equal(suite.T(), rentalBefore.State, rental.State)
	assert.Equal(suite.T(), rentalBefore.RentalPeriod, rental.RentalPeriod)
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_leftOverlap() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	leftOverlapTimePeriod := model.TimePeriod{
		StartDate: timePeriod.StartDate.Add(-24 * time.Hour),
		EndDate:   timePeriod.EndDate.Add(-24 * time.Hour),
	}

	timePeriod.EndDate = leftOverlapTimePeriod.EndDate

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(leftOverlapTimePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_success_rightOverlap() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rightOverlapTimePeriod := model.TimePeriod{
		StartDate: timePeriod.StartDate.Add(24 * time.Hour),
		EndDate:   timePeriod.EndDate.Add(24 * time.Hour),
	}

	timePeriod.StartDate = rightOverlapTimePeriod.StartDate

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(rightOverlapTimePeriod).
		Expect(suite.T()).
		Status(http.StatusCreated).
		Assert(suite.assertToken(timePeriod)).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_invalidRentalId() {
	suite.newApiTestWithCarMock().
		Post("/rentals/waytoolongrentalid/trunkTokens").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_invalidTimePeriod() {
	suite.newApiTestWithCarMock().
		Post("/rentals/rentalId/trunkTokens").
		JSON(testdata.TimePeriodSemanticInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_invalidTimePeriod2() {
	suite.newApiTestWithCarMock().
		Post("/rentals/rentalId/trunkTokens").
		JSON(testdata.TimePeriodSyntaxInvalid).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_missingTimePeriod() {
	suite.newApiTestWithCarMock().
		Post("/rentals/rentalId/trunkTokens").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_upcoming() {
	suite.createRentalForCustomer(testdata.VinCar, testdata.TimePeriod2122, "customer")

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_expired() {
	now := time.Now().UTC().Round(time.Millisecond)
	timePeriod := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Millisecond),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(15 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(marshalledTime).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_notOverlappingFuture() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(testdata.TimePeriod2150).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_notOverlappingPast() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	suite.newApiTestWithCarMock().
		Post("/rentals/" + rentalId + "/trunkTokens").
		JSON(testdata.TimePeriod1900).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGrantTrunkAccess_unknownRentalId() {
	suite.newApiTestWithCarMock().
		Post("/rentals/unknown1/trunkTokens").
		JSON(testdata.TimePeriod2122).
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}

func returnsRental(expectedRental model.Rental, t *testing.T) func(*http.Response, *http.Request) error {
	return func(res *http.Response, _ *http.Request) error {
		defer func() { _ = res.Body.Close() }()

		var rental model.Rental
		if err := json.NewDecoder(res.Body).Decode(&rental); err != nil {
			return err
		}

		rental.Id = ""

		assert.Equal(t, expectedRental, rental)

		return nil
	}
}

func (suite *ApiTestSuite) TestGetNextRental_success_existsInFuture() {
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)
	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRental(model.Rental{
			State: model.UPCOMING,
			RentalPeriod: model.TimePeriod{
				EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
				StartDate: time.Date(2122, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Customer: &model.Customer{
				CustomerId: "d9ChwOvI",
			},
		}, suite.T())).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_success_existsInPastAndFuture() {
	now := time.Now().UTC().Round(time.Millisecond)
	periodForPastRental := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Millisecond),
	}
	marshalledPeriodForPastRental, _ := json.Marshal(periodForPastRental)

	suite.createRental(testdata.VinCar, string(marshalledPeriodForPastRental))
	suite.createRental(testdata.VinCar, testdata.TimePeriod2122)

	time.Sleep(15*time.Millisecond - time.Since(now))

	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRental(model.Rental{
			State: model.UPCOMING,
			RentalPeriod: model.TimePeriod{
				EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
				StartDate: time.Date(2122, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Customer: &model.Customer{
				CustomerId: "d9ChwOvI",
			},
		}, suite.T())).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_success_existsNow() {
	now := time.Now().UTC().Round(time.Millisecond)
	periodForPastRental := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Millisecond),
	}
	marshalledPeriodForPastRental, _ := json.Marshal(periodForPastRental)

	suite.createRental(testdata.VinCar, string(marshalledPeriodForPastRental))

	now = time.Now().UTC().Round(time.Millisecond)
	periodFromNow := model.TimePeriod{
		StartDate: now.Add(13 * time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	marshalledPeriodFromNow, _ := json.Marshal(periodFromNow)

	suite.createRental(testdata.VinCar, string(marshalledPeriodFromNow))

	suite.createRental(testdata.VinCar, testdata.TimePeriod2150)

	time.Sleep(15*time.Millisecond - time.Since(now))

	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusOK).
		Assert(returnsRental(model.Rental{
			State:        model.ACTIVE,
			RentalPeriod: periodFromNow,
			Customer: &model.Customer{
				CustomerId: "d9ChwOvI",
			},
		}, suite.T())).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_success_noRentalExists() {
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusNoContent).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_success_onlyRentalsInPast() {
	now := time.Now().UTC().Round(time.Millisecond)
	periodFromNow := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond),
		EndDate:   now.Add(12 * time.Millisecond),
	}
	marshalledPeriodFromNow, _ := json.Marshal(periodFromNow)

	suite.createRental(testdata.VinCar, string(marshalledPeriodFromNow))

	time.Sleep(15*time.Millisecond - time.Since(now))

	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusNoContent).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_invalidVin() {
	suite.newApiTestWithCarMock().
		Get("/cars/invalidVin/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetNextRental_carDoesNotExist() {
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.UnknownVin + "/rentalStatus").
		Expect(suite.T()).
		Status(http.StatusNotFound).
		End()
}
func (suite *ApiTestSuite) TestGetLockState_success() {
	timePeriodRental := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	timePeriodAccess := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	token := suite.grantTrunkAccess(rentalId, timePeriodAccess).Token

	time.Sleep(10 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		Expect(suite.T()).
		Status(http.StatusOK).
		Body(testdata.Unlocked).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_useTokenWithOtherCar() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriod).Token

	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar2+"/trunk").
		Query("trunkAccessToken", token).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_trunkAccessTokenExpired() {

	now := time.Now().UTC().Round(time.Millisecond)

	timePeriodRental := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	now = time.Now().UTC().Round(time.Millisecond)

	timePeriodAccess := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   now.Add(15 * time.Millisecond).UTC().Round(time.Millisecond),
	}

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriodAccess).Token

	time.Sleep(15 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_trunkAccessTokenValidInFuture() {
	now := time.Now().UTC().Round(time.Millisecond)

	timePeriodRental := model.TimePeriod{
		StartDate: now.Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	now = time.Now().UTC().Round(time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	timePeriodFutureInRental := model.TimePeriod{
		StartDate: time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2900, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	token := suite.grantTrunkAccess(rentalId, timePeriodFutureInRental).Token

	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_unknownTrunkAccessToken() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriod).Token
	if strings.Contains(token, "b") {
		token = strings.Replace(token, "b", "a", 1)
	} else {
		token = testdata.TrunkAccessToken
	}
	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_noTrunkAccessToken() {
	suite.newApiTestWithCarMock().
		Get("/cars/" + testdata.VinCar + "/trunk").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_invalidTrunkAccessToken() {
	suite.newApiTestWithCarMock().
		Get("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", "invalidTrunkAccessToken").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestGetLockState_invalidVin() {
	suite.newApiTestWithCarMock().
		Get("/cars/invalidVin/trunk").
		Query("trunkAccessToken", testdata.TrunkAccessToken).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_success() {
	periodFromNow := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledPeriodFromNow, _ := json.Marshal(periodFromNow)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledPeriodFromNow), "d9ChwOvI")

	time.Sleep(10 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusNoContent).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_noParameters() {
	suite.newApiTestWithCarMock().
		Put("/cars/" + testdata.VinCar + "/trunk").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_tooManyParameters() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "d9ChwOvI").
		Query("trunkAccessToken", testdata.TrunkAccessToken).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_noRentalsOfCustomer() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_rentalUpcoming() {
	suite.createRentalForCustomer(testdata.VinCar, testdata.TimePeriod2123, "d9ChwOvI")

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_rentalExpired() {
	timePeriodRental := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Now().Add(15 * time.Millisecond).UTC().Round(time.Millisecond),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(15 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "customer").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_invalidVin() {
	suite.newApiTestWithCarMock().
		Put("/cars/invalidVin/trunk").
		Query("customerId", "d9ChwOvI").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_invalidCustomerId() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "invalidCustomerId").
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_customerId_invalidLockState() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("customerId", "d9ChwOvI").
		JSON("").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_success() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriod).Token

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusNoContent).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_tokenNotFound() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", testdata.TrunkAccessToken).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_tokenValidForOtherCar() {
	timePeriod := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriod)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriod).Token

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar2+"/trunk").
		Query("trunkAccessToken", token).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_tokenValidInPast() {
	timePeriodRental := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(2123, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	rentalId := suite.getRentalOverview("customer")[0].Id

	timePeriodAccess := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Now().Add(11 * time.Millisecond).UTC().Round(time.Millisecond),
	}

	token := suite.grantTrunkAccess(rentalId, timePeriodAccess).Token

	time.Sleep(20 * time.Millisecond)

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_tokenValidInFuture() {
	timePeriodRental := model.TimePeriod{
		StartDate: time.Now().Add(10 * time.Millisecond).UTC().Round(time.Millisecond),
		EndDate:   time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	marshalledTime, _ := json.Marshal(timePeriodRental)

	suite.createRentalForCustomer(testdata.VinCar, string(marshalledTime), "customer")

	time.Sleep(10 * time.Millisecond)

	timePeriodFutureInRental := model.TimePeriod{
		StartDate: time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2900, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	rentalId := suite.getRentalOverview("customer")[0].Id

	token := suite.grantTrunkAccess(rentalId, timePeriodFutureInRental).Token

	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", token).
		JSON(testdata.Locked).
		Expect(suite.T()).
		Status(http.StatusForbidden).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_invalidVin() {
	suite.newApiTestWithCarMock().
		Put("/cars/invalidVin/trunk").
		Query("trunkAccessToken", testdata.TrunkAccessToken).
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}

func (suite *ApiTestSuite) TestSetLockState_trunkAccessToken_invalidLockState() {
	suite.newApiTestWithCarMock().
		Put("/cars/"+testdata.VinCar+"/trunk").
		Query("trunkAccessToken", testdata.TrunkAccessToken).
		JSON("").
		Expect(suite.T()).
		Status(http.StatusBadRequest).
		End()
}
