package main

import (
	"RentalManagement/api"
	"RentalManagement/environment"
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/logic/operations"
	"RentalManagement/util"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
)

// newApp allows production as well as testing to create a new Echo instance for the API
// Configuration values are read from the environment.
func newApp(dbConnection db.IConnection) (*echo.Echo, error) {
	app := echo.New()

	// add CORS middleware if allowed origins are configured
	allowOrigins := environment.GetEnvironment().GetAppAllowOrigins()
	if len(allowOrigins) > 0 {
		app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
		}))
	}

	// add OpenAPI validation to the echo instance
	err := api.AddOpenApiValidationMiddleware(app)
	if err != nil {
		return nil, err
	}

	carClient, err := car.NewClientWithResponses(environment.GetEnvironment().GetCarServerUrl(),
		car.WithHTTPClient(
			&http.Client{
				Timeout: environment.GetEnvironment().GetRequestTimeout(),
			},
		),
	)

	if err != nil {
		return nil, err
	}

	crudInstance := database.NewICRUD(dbConnection, environment.GetEnvironment(), util.TimeProvider{})
	operationsInstance := operations.NewOperations(carClient, crudInstance, util.TimeProvider{})
	controllerInstance := api.NewController(operationsInstance, util.TimeProvider{})

	api.RegisterHandlers(app, controllerInstance)

	// Use custom error handling that logs any rentalErrors that occur but passes any HTTP rentalErrors directly to the client.
	// Any other rentalErrors are converted to HTTP 500 rentalErrors.
	app.Use(func(fun echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := fun(c); err != nil {
				if err, isHttpError := err.(*echo.HTTPError); isHttpError {
					return err
				}
				app.Logger.Error(err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
			}
			return nil
		}
	})

	return app, nil
}

func main() {
	dbConnection, err := db.NewDbConnection(environment.GetEnvironment())
	if err != nil {
		log.Fatal(err)
	}

	app, err := newApp(dbConnection)
	if err != nil {
		log.Fatal(err)
	}

	// start the server on the configured port
	app.Logger.Fatal(app.Start(fmt.Sprintf(":%d", environment.GetEnvironment().GetAppExposePort())))
}
