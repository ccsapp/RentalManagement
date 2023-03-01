package environment

import (
	_ "embed"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envMongoDbHost         = "MONGODB_DATABASE_HOST"
	envMongoDbPort         = "MONGODB_DATABASE_PORT"
	envMongoDbDatabase     = "MONGODB_DATABASE_NAME"
	envMongoDbUser         = "MONGODB_DATABASE_USER"
	envMongoDbPassword     = "MONGODB_DATABASE_PASSWORD"
	envAppExposePort       = "RM_EXPOSE_PORT"
	envAppCollectionPrefix = "RM_COLLECTION_PREFIX"
	envCarServerUrl        = "RM_CAR_SERVER"
	envRequestTimeout      = "RM_REQUEST_TIMEOUT"
	envAppAllowOrigins     = "RM_ALLOW_ORIGINS"
	envLocalSetupMode      = "RM_LOCAL_SETUP"

	defaultMongoDbPort         = 27017
	defaultAppExposePort       = 80
	defaultAppCollectionPrefix = ""
	defaultRequestTimeout      = 5 * time.Second
)

var defaultAppAllowOrigins []string

func ptr[T any](v T) *T {
	return &v
}

//go:embed localSetup.env
var localSetup string

// readEnvironment reads the correct environment configuration (also considering local setup mode)
func readEnvironment() *Environment {
	if getBooleanEnvVariable(envLocalSetupMode) {
		fmt.Println("Using local setup mode.")

		localSetupMap, err := godotenv.Unmarshal(localSetup)
		if err != nil {
			panic("Invalid local setup environment variables. This is a bug.")
		}

		// Unfortunately, godotenv does not support reading environment variables from a string
		// directly to the environment. Therefore, we have to use this workaround.
		for key, value := range localSetupMap {
			if os.Getenv(key) != "" {
				// do not overwrite existing environment variables
				continue
			}
			_ = os.Setenv(key, value)
		}
	}
	return readEnvironmentFromEnv()
}

// readEnvironmentFromEnv reads the environment configuration from actual environment variables
// If any of the required environment variables is not set, the program will panic.
func readEnvironmentFromEnv() *Environment {
	return &Environment{
		mongoDbHost:         getStringEnvVariable(envMongoDbHost, nil),
		mongoDbPort:         getIntegerEnvVariable(envMongoDbPort, ptr(defaultMongoDbPort)),
		mongoDbDatabase:     getStringEnvVariable(envMongoDbDatabase, nil),
		mongoDbUser:         getStringEnvVariable(envMongoDbUser, nil),
		mongoDbPassword:     getStringEnvVariable(envMongoDbPassword, nil),
		appExposePort:       getIntegerEnvVariable(envAppExposePort, ptr(defaultAppExposePort)),
		appCollectionPrefix: getStringEnvVariable(envAppCollectionPrefix, ptr(defaultAppCollectionPrefix)),
		carServerUrl:        getStringEnvVariable(envCarServerUrl, nil),
		requestTimeout:      getDurationEnvVariable(envRequestTimeout, ptr(defaultRequestTimeout)),
		appAllowOrigins:     getStringArrayEnvVariable(envAppAllowOrigins, ptr(defaultAppAllowOrigins)),
		isLocalSetupMode:    getBooleanEnvVariable(envLocalSetupMode),
	}
}

// getStringEnvVariable returns the string value of the environment variable with the given name.
// You can specify a default value that is returned if the environment variable is not set,
// set defaultValue to nil to disable this feature.
// If defaultValue is nil and the environment variable is not set, the program will panic.
func getStringEnvVariable(variableName string, defaultValue *string) string {
	stringValue := os.Getenv(variableName)
	if stringValue != "" {
		return stringValue
	}

	if defaultValue != nil {
		return *defaultValue
	}
	panic("Environment variable not set: " + variableName)
}

// getIntegerEnvVariable returns the integer value of the environment variable with the given name.
// You can specify a default value that is returned if the environment variable is not set,
// set defaultValue to nil to disable this feature.
// If defaultValue is nil and the environment variable is not set, the program will panic.
// If the environment variable is not a valid integer value, the program will panic.
func getIntegerEnvVariable(variableName string, defaultValue *int) int {
	var stringValue string
	if defaultValue != nil {
		defaultValueString := strconv.Itoa(*defaultValue)
		stringValue = getStringEnvVariable(variableName, &defaultValueString)
	} else {
		stringValue = getStringEnvVariable(variableName, nil)
	}

	intValue, err := strconv.Atoi(stringValue)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for integer environment variable \"%s\": %s",
			variableName, stringValue))
	}
	return intValue
}

// getBooleanEnvVariable returns the boolean value of the environment variable with the given name.
// If the environment variable is not set, false is returned.
// If the environment variable is not a valid boolean value, the program will panic.
func getBooleanEnvVariable(variableName string) bool {
	stringValue := os.Getenv(variableName)
	if stringValue == "" || stringValue == "false" {
		return false
	}

	if stringValue == "true" {
		return true
	}

	panic(fmt.Sprintf("Invalid value for boolean environment variable \"%s\": %s",
		variableName, stringValue))
}

// getDurationEnvVariable returns the duration value of the environment variable with the given name.
// You can specify a default value that is returned if the environment variable is not set,
// set defaultValue to nil to disable this feature.
// If defaultValue is nil and the environment variable is not set, the program will panic.
// If the environment variable is not a valid duration value, the program will panic.
func getDurationEnvVariable(variableName string, defaultValue *time.Duration) time.Duration {
	var stringValue string
	if defaultValue != nil {
		defaultValueString := defaultValue.String()
		stringValue = getStringEnvVariable(variableName, &defaultValueString)
	} else {
		stringValue = getStringEnvVariable(variableName, nil)
	}

	durationValue, err := time.ParseDuration(stringValue)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for duration environment variable \"%s\": %s",
			variableName, stringValue))
	}
	return durationValue
}

// getStringArrayEnvVariable returns the string array value of the environment variable with the given name.
// The string array value is parsed from a comma-separated string.
// You can specify a default value that is returned if the environment variable is not set.
// nil (empty slice) is supported as default value but not as environment variable value.
// If the environment variable is not a valid string array value, the program will panic.
func getStringArrayEnvVariable(variableName string, defaultValue *[]string) []string {
	defaultValueString := strings.Join(*defaultValue, ",")
	stringValue := getStringEnvVariable(variableName, &defaultValueString)

	if stringValue == "" {
		// empty slice
		return []string{}
	}

	return strings.Split(stringValue, ",")
}
