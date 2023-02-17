package main

import (
	"RentalManagement/api"
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/logic/operations"
	"RentalManagement/util"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	EnvAllowOrigins  = "RM_ALLOW_ORIGINS"
	EnvDomainServer  = "RM_DOMAIN_SERVER"
	EnvDomainTimeout = "RM_DOMAIN_TIMEOUT"
)

type Config struct {
	allowOrigins  []string
	domainServer  string
	domainTimeout time.Duration
}

// newApp allows production as well as testing to create a new Echo instance for the API
func newApp(config *Config, dbConnection db.IConnection, dbConfig *db.Config) (*echo.Echo, error) {
	app := echo.New()

	if len(config.allowOrigins) > 0 {
		app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: config.allowOrigins,
		}))
	}

	// add OpenAPI validation to the echo instance
	err := api.AddOpenApiValidationMiddleware(app)
	if err != nil {
		return nil, err
	}

	carClient, err := car.NewClientWithResponses(config.domainServer,
		car.WithHTTPClient(
			&http.Client{
				Timeout: config.domainTimeout,
			},
		),
	)

	if err != nil {
		return nil, err
	}

	crudInstance := database.NewICRUD(dbConnection, dbConfig, util.TimeProvider{})
	operationsInstance := operations.NewOperations(carClient, crudInstance)
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

func loadConfig() (*Config, error) {
	allowOriginsString := os.Getenv(EnvAllowOrigins)
	var allowOrigins []string
	if allowOriginsString != "" {
		allowOrigins = strings.Split(allowOriginsString, ",")
	} else {
		allowOrigins = []string{}
	}

	domainServer := os.Getenv(EnvDomainServer)
	if domainServer == "" {
		return nil, errors.New("no domain server given")
	}

	timeoutString := os.Getenv(EnvDomainTimeout)
	var domainTimeout time.Duration

	if timeoutString != "" {
		var err error // declaring with := below would create separate domainTimeout var in this scope
		domainTimeout, err = time.ParseDuration(timeoutString)
		if err != nil {
			return nil, errors.New("invalid domain timeout configured")
		}
	} else {
		domainTimeout = 5 * time.Second
	}

	return &Config{
		allowOrigins,
		domainServer,
		domainTimeout,
	}, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbConfig, err := db.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	dbConnection, err := db.NewDbConnection(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	app, err := newApp(config, dbConnection, dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	app.Logger.Fatal(app.Start(":80"))
}
