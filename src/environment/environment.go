package environment

import "time"

var (
	environment *Environment
)

// GetEnvironment returns the environment configuration.
// If the environment configuration has not been read yet, it will be read from the environment variables.
// If any error occurs while reading the environment configuration, the program will panic.
// If the environment configuration has already been read, the cached value will be returned.
func GetEnvironment() *Environment {
	if environment == nil {
		environment = readEnvironment()
	}
	return environment
}

// SetupTestingEnvironment sets up the testing environment configuration.
func SetupTestingEnvironment(carServerUrl string) {
	setCarServerUrl(carServerUrl)
	environment = readEnvironment()
}

type Environment struct {
	mongoDbConnectionString string
	mongoDbDatabase         string
	appExposePort           int
	appCollectionPrefix     string
	carServerUrl            string
	requestTimeout          time.Duration
	appAllowOrigins         []string
	isLocalSetupMode        bool
}

func (e *Environment) GetMongoDbConnectionString() string {
	return e.mongoDbConnectionString
}

func (e *Environment) GetMongoDbDatabase() string {
	return e.mongoDbDatabase
}

func (e *Environment) GetAppExposePort() int {
	return e.appExposePort
}

func (e *Environment) GetAppCollectionPrefix() string {
	return e.appCollectionPrefix
}

// SetAppCollectionPrefix sets the prefix for collection names of the application.
// This method should only be used for testing.
func (e *Environment) SetAppCollectionPrefix(prefix string) {
	e.appCollectionPrefix = prefix
}

func (e *Environment) GetCarServerUrl() string {
	return e.carServerUrl
}

func (e *Environment) GetRequestTimeout() time.Duration {
	return e.requestTimeout
}

func (e *Environment) GetAppAllowOrigins() []string {
	return e.appAllowOrigins
}

func (e *Environment) IsLocalSetupMode() bool {
	return e.isLocalSetupMode
}
