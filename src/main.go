package main

import (
	"RentalManagement/api"
	"RentalManagement/infrastructure/car"
	"RentalManagement/logic/operations"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e := echo.New()

	config, err := loadConfig()
	if err != nil {
		e.Logger.Fatal(err)
	}

	if len(config.allowOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: config.allowOrigins,
		}))
	}

	carClient, err := car.NewClientWithResponses(config.domainServer,
		car.WithHTTPClient(
			&http.Client{
				Timeout: config.domainTimeout,
			},
		),
	)

	if err != nil {
		e.Logger.Fatal(err)
	}

	operationsInstance := operations.NewOperations(carClient)
	controllerInstance := api.NewController(operationsInstance)

	api.RegisterHandlers(e, controllerInstance)

	e.Logger.Fatal(e.Start(":80"))
}
