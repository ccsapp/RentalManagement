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

func newApiTest(handler http.Handler, name string) *apitest.APITest {
	return apitest.New(name).
		Debug().
		Handler(handler).
		Report(apitest.SequenceDiagram())
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
			RespondWith().Status(http.StatusOK).Body(testdata.ExampleCar).End(),
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars/" + testdata.VinCar2).
			RespondWith().Status(http.StatusOK).Body(testdata.ExampleCar2).End(),
		apitest.NewMock().
			Get(suite.config.domainServer + "/cars/" + testdata.UnknownVin).
			RespondWith().Status(http.StatusNotFound).End(),
	}
}
