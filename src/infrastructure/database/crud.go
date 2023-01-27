package database

//go:generate mockgen -source=./crud.go -package=mocks -destination=../../mocks/mock_crud.go

import (
	"RentalManagement/infrastructure/database/db"
	"RentalManagement/util"
)

const CollectionBaseName = "rentals"

// ICRUD is a high level database interface. It directly maps to the business logic and abstracts away the
// database entities and the database connection.
type ICRUD interface {
}

type crud struct {
	db           db.IConnection
	collection   string
	timeProvider util.TimeProvider
}

func NewICRUD(db db.IConnection, config *db.Config, provider util.TimeProvider) ICRUD {
	return &crud{
		db:           db,
		collection:   config.CollectionPrefix + CollectionBaseName,
		timeProvider: provider,
	}
}
