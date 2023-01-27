// Package car provides primitives to interact with the openapi HTTP API of domain microservice Car.
package car

//go:generate mockgen -source=client.go -package=mocks -destination=../../mocks/mock_car_client.go

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	carTypes "git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-cargotypes.git"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
)

// HttpRequestDoer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// NewClient creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// ClientInterface specification for the client above.
type ClientInterface interface {
	// GetCars get the VINs of all cars maintained by the system
	GetCars(ctx context.Context) (*http.Response, error)

	// GetCar get all information about a specific car
	GetCar(ctx context.Context, vin carTypes.VinParam) (*http.Response, error)

	// ChangeTrunkLockState open or close trunk
	ChangeTrunkLockState(ctx context.Context, vin carTypes.VinParam, body carTypes.DynamicDataLockState) (*http.Response, error)
}

func (c *Client) GetCars(ctx context.Context) (*http.Response, error) {
	req, err := NewGetCarsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

func (c *Client) GetCar(ctx context.Context, vin carTypes.VinParam) (*http.Response, error) {
	req, err := NewGetCarRequest(c.Server, vin)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

func (c *Client) ChangeTrunkLockState(ctx context.Context, vin carTypes.VinParam, body carTypes.DynamicDataLockState) (*http.Response, error) {
	req, err := NewChangeTrunkLockStateRequest(c.Server, vin, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// NewGetCarsRequest generates requests for GetCars
func NewGetCarsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/cars")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetCarRequest generates requests for GetCar
func NewGetCarRequest(server string, vin carTypes.VinParam) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "vin", runtime.ParamLocationPath, vin)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/cars/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewChangeTrunkLockStateRequest calls the generic ChangeTrunkLockState builder with application/json body
func NewChangeTrunkLockStateRequest(server string, vin carTypes.VinParam, body carTypes.DynamicDataLockState) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return newChangeTrunkLockStateRequestWithBody(server, vin, "application/json", bodyReader)
}

// newChangeTrunkLockStateRequestWithBody generates requests for ChangeTrunkLockState with any type of body
func newChangeTrunkLockStateRequestWithBody(server string, vin carTypes.VinParam, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "vin", runtime.ParamLocationPath, vin)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/cars/%s/trunkLock", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetCarsWithResponse get the VINs of all cars maintained by the system
	GetCarsWithResponse(ctx context.Context) (*GetCarsResponse, error)

	// GetCarWithResponse get all information about a specific car
	GetCarWithResponse(ctx context.Context, vin carTypes.VinParam) (*GetCarResponse, error)

	// ChangeTrunkLockStateWithResponse open or close trunk
	ChangeTrunkLockStateWithResponse(ctx context.Context, vin carTypes.VinParam, body carTypes.DynamicDataLockState) (*ChangeTrunkLockStateResponse, error)
}

type GetCarsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	ParsedVins   *[]carTypes.Vin
}

// Status returns HTTPResponse.Status
func (r GetCarsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetCarsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetCarResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	ParsedCar    *carTypes.Car
}

// Status returns HTTPResponse.Status
func (r GetCarResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetCarResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ChangeTrunkLockStateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r ChangeTrunkLockStateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ChangeTrunkLockStateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetCarsWithResponse request returning *GetCarsResponse
func (c *ClientWithResponses) GetCarsWithResponse(ctx context.Context) (*GetCarsResponse, error) {
	rsp, err := c.GetCars(ctx)
	if err != nil {
		return nil, err
	}
	return ParseGetCarsResponse(rsp)
}

// GetCarWithResponse request returning *GetCarResponse
func (c *ClientWithResponses) GetCarWithResponse(ctx context.Context, vin carTypes.VinParam) (*GetCarResponse, error) {
	rsp, err := c.GetCar(ctx, vin)
	if err != nil {
		return nil, err
	}
	return ParseGetCarResponse(rsp)
}

func (c *ClientWithResponses) ChangeTrunkLockStateWithResponse(ctx context.Context, vin carTypes.VinParam, body carTypes.DynamicDataLockState) (*ChangeTrunkLockStateResponse, error) {
	rsp, err := c.ChangeTrunkLockState(ctx, vin, body)
	if err != nil {
		return nil, err
	}
	return ParseChangeTrunkLockStateResponse(rsp)
}

// ParseGetCarsResponse parses an HTTP response from a GetCarsWithResponse call
func ParseGetCarsResponse(rsp *http.Response) (*GetCarsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetCarsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []carTypes.Vin
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ParsedVins = &dest

	}

	return response, nil
}

// ParseGetCarResponse parses an HTTP response from a GetCarWithResponse call
func ParseGetCarResponse(rsp *http.Response) (*GetCarResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetCarResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest carTypes.Car
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ParsedCar = &dest

	}

	return response, nil
}

// ParseChangeTrunkLockStateResponse parses an HTTP response from a ChangeTrunkLockStateWithResponse call
func ParseChangeTrunkLockStateResponse(rsp *http.Response) (*ChangeTrunkLockStateResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ChangeTrunkLockStateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}
