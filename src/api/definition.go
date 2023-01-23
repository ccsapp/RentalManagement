// Package api provides primitives to interact with the openapi HTTP API.
package api

import (
	"RentalManagement/logic/model"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// GetAvailableCars Get Available Cars in a Time Period
	// (GET /cars)
	GetAvailableCars(ctx echo.Context, params model.GetAvailableCarsParams) error
	// GetCar Get Static Information On a Car
	// (GET /cars/{vin})
	GetCar(ctx echo.Context, vin model.VinParam) error
	// GetNextRental Get the Active or Next Upcoming Rental
	// (GET /cars/{vin}/rentalStatus)
	GetNextRental(ctx echo.Context, vin model.VinParam) error
	// CreateRental Create a New Rental
	// (POST /cars/{vin}/rentals)
	CreateRental(ctx echo.Context, vin model.VinParam, params model.CreateRentalParams) error
	// GetLockState Get the Trunk Lock State of the Car
	// (GET /cars/{vin}/trunk)
	GetLockState(ctx echo.Context, vin model.VinParam, params model.GetLockStateParams) error
	// SetLockState Set the Trunk Lock State of the Car
	// (PUT /cars/{vin}/trunk)
	SetLockState(ctx echo.Context, vin model.VinParam, params model.SetLockStateParams) error
	// GetOverview Get an Overview of a Customer’s Rentals
	// (GET /rentals)
	GetOverview(ctx echo.Context, params model.GetOverviewParams) error
	// GetRentalStatus Get the Status of the Rental and the Car
	// (GET /rentals/{rentalId})
	GetRentalStatus(ctx echo.Context, rentalId model.RentalIdParam) error
	// GrantTrunkAccess Create a New Token to Access the Trunk
	// (POST /rentals/{rentalId}/trunkTokens)
	GrantTrunkAccess(ctx echo.Context, rentalId model.RentalIdParam) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetAvailableCars converts echo context to params.
func (w *ServerInterfaceWrapper) GetAvailableCars(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params model.GetAvailableCarsParams
	// ------------- Required query parameter "timePeriod" -------------

	err = runtime.BindQueryParameter("form", true, true, "timePeriod", ctx.QueryParams(), &params.TimePeriod)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter timePeriod: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetAvailableCars(ctx, params)
	return err
}

// GetCar converts echo context to params.
func (w *ServerInterfaceWrapper) GetCar(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "vin" -------------
	var vin model.VinParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "vin", runtime.ParamLocationPath, ctx.Param("vin"), &vin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter vin: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetCar(ctx, vin)
	return err
}

// GetNextRental converts echo context to params.
func (w *ServerInterfaceWrapper) GetNextRental(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "vin" -------------
	var vin model.VinParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "vin", runtime.ParamLocationPath, ctx.Param("vin"), &vin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter vin: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetNextRental(ctx, vin)
	return err
}

// CreateRental converts echo context to params.
func (w *ServerInterfaceWrapper) CreateRental(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "vin" -------------
	var vin model.VinParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "vin", runtime.ParamLocationPath, ctx.Param("vin"), &vin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter vin: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params model.CreateRentalParams
	// ------------- Required query parameter "customerId" -------------

	err = runtime.BindQueryParameter("form", true, true, "customerId", ctx.QueryParams(), &params.CustomerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter customerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateRental(ctx, vin, params)
	return err
}

// GetLockState converts echo context to params.
func (w *ServerInterfaceWrapper) GetLockState(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "vin" -------------
	var vin model.VinParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "vin", runtime.ParamLocationPath, ctx.Param("vin"), &vin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter vin: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params model.GetLockStateParams
	// ------------- Required query parameter "trunkAccessToken" -------------

	err = runtime.BindQueryParameter("form", true, true, "trunkAccessToken", ctx.QueryParams(), &params.TrunkAccessToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trunkAccessToken: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetLockState(ctx, vin, params)
	return err
}

// SetLockState converts echo context to params.
func (w *ServerInterfaceWrapper) SetLockState(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "vin" -------------
	var vin model.VinParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "vin", runtime.ParamLocationPath, ctx.Param("vin"), &vin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter vin: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params model.SetLockStateParams
	// ------------- Optional query parameter "customerId" -------------

	err = runtime.BindQueryParameter("form", true, false, "customerId", ctx.QueryParams(), &params.CustomerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter customerId: %s", err))
	}

	// ------------- Optional query parameter "trunkAccessToken" -------------

	err = runtime.BindQueryParameter("form", true, false, "trunkAccessToken", ctx.QueryParams(), &params.TrunkAccessToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trunkAccessToken: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SetLockState(ctx, vin, params)
	return err
}

// GetOverview converts echo context to params.
func (w *ServerInterfaceWrapper) GetOverview(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params model.GetOverviewParams
	// ------------- Required query parameter "customerId" -------------

	err = runtime.BindQueryParameter("form", true, true, "customerId", ctx.QueryParams(), &params.CustomerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter customerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetOverview(ctx, params)
	return err
}

// GetRentalStatus converts echo context to params.
func (w *ServerInterfaceWrapper) GetRentalStatus(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "rentalId" -------------
	var rentalId model.RentalIdParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "rentalId", runtime.ParamLocationPath, ctx.Param("rentalId"), &rentalId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter rentalId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetRentalStatus(ctx, rentalId)
	return err
}

// GrantTrunkAccess converts echo context to params.
func (w *ServerInterfaceWrapper) GrantTrunkAccess(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "rentalId" -------------
	var rentalId model.RentalIdParam

	err = runtime.BindStyledParameterWithLocation("simple", false, "rentalId", runtime.ParamLocationPath, ctx.Param("rentalId"), &rentalId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter rentalId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GrantTrunkAccess(ctx, rentalId)
	return err
}

// EchoRouter is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// RegisterHandlersWithBaseURL registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/cars", wrapper.GetAvailableCars)
	router.GET(baseURL+"/cars/:vin", wrapper.GetCar)
	router.GET(baseURL+"/cars/:vin/rentalStatus", wrapper.GetNextRental)
	router.POST(baseURL+"/cars/:vin/rentals", wrapper.CreateRental)
	router.GET(baseURL+"/cars/:vin/trunk", wrapper.GetLockState)
	router.PUT(baseURL+"/cars/:vin/trunk", wrapper.SetLockState)
	router.GET(baseURL+"/rentals", wrapper.GetOverview)
	router.GET(baseURL+"/rentals/:rentalId", wrapper.GetRentalStatus)
	router.POST(baseURL+"/rentals/:rentalId/trunkTokens", wrapper.GrantTrunkAccess)

}
