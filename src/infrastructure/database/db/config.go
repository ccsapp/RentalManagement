package db

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
)

const (
	envDbHost       = "MONGODB_DATABASE_HOST"
	envDbDatabase   = "MONGODB_DATABASE_NAME"
	envDbUser       = "MONGODB_DATABASE_USER"
	envDbPassword   = "MONGODB_DATABASE_PASSWORD"
	envDbCollPrefix = "RENTAL_MANAGEMENT_COLLECTION_PREFIX"
)

var (
	errIncompleteConfig = errors.New("the configuration is incomplete")
)

type Config struct {
	Host             string
	Db               string
	User             string
	Password         string
	CollectionPrefix string
}

// LoadConfigFromEnv loads the database configuration from environment variables with predefined names
func LoadConfigFromEnv() (*Config, error) {
	// os.Getenv returns empty string if variable not set

	host := os.Getenv(envDbHost)
	if host == "" {
		return nil, errIncompleteConfig
	}

	db := os.Getenv(envDbDatabase)
	if db == "" {
		return nil, errIncompleteConfig
	}

	user := os.Getenv(envDbUser)
	if user == "" {
		return nil, errIncompleteConfig
	}

	password := os.Getenv(envDbPassword)
	if password == "" {
		return nil, errIncompleteConfig
	}

	// may not be set => results in empty prefix which is fine
	collectionPrefix := os.Getenv(envDbCollPrefix)

	return &Config{
		host,
		db,
		user,
		password,
		collectionPrefix,
	}, nil
}

// LoadConfigFromFile loads the database configuration from a file in env syntax
func LoadConfigFromFile(filename string) (*Config, error) {
	envMap, err := godotenv.Read(filename)
	if err != nil {
		return nil, err
	}

	var host, db, user, password, collectionPrefix string
	var ok bool

	if host, ok = envMap[envDbHost]; !ok {
		return nil, errIncompleteConfig
	}

	if db, ok = envMap[envDbDatabase]; !ok {
		return nil, errIncompleteConfig
	}

	if user, ok = envMap[envDbUser]; !ok {
		return nil, errIncompleteConfig
	}

	if password, ok = envMap[envDbPassword]; !ok {
		return nil, errIncompleteConfig
	}

	if collectionPrefix, ok = envMap[envDbCollPrefix]; !ok {
		collectionPrefix = "" // default is no (= empty) prefix
	}

	return &Config{
		host,
		db,
		user,
		password,
		collectionPrefix,
	}, nil
}
