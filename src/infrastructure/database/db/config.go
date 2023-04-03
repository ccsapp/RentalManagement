package db

type DatabaseConfig interface {
	GetMongoDbConnectionString() string
	GetMongoDbDatabase() string
}
