package api

import (
	_ "embed"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

//go:embed openapi.yaml
var openApiData []byte

// AddOpenApiValidationMiddleware adds validation middleware to the echo server. It uses the OpenAPI specification of
// RentalManagement to validate API requests.
func AddOpenApiValidationMiddleware(e *echo.Echo) error {
	swagger, err := openapi3.NewLoader().LoadFromData(openApiData)
	if err != nil {
		return err
	}

	e.Use(middleware.OapiRequestValidator(swagger))

	return nil
}
