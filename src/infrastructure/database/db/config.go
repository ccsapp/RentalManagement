package db

type DatabaseConfig interface {
	GetMongoDbHost() string
	GetMongoDbPort() int
	GetMongoDbDatabase() string
	GetMongoDbUser() string
	GetMongoDbPassword() string
}
