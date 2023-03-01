package db

//go:generate mockgen -source=./connection.go -package=mocks -destination=../../../mocks/mock_connection.go

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// IConnection is a low level interface to the database.
type IConnection interface {
	// CleanUpDatabase closes the connection to the database. You should defer a call to this method after calling
	// NewDbConnection. Any rentalErrors are unexpected.
	CleanUpDatabase() error

	// GetFactory gets a factory to create queries for this connection.
	GetFactory() QueryFactory

	// Insert inserts a document into the database at the specified collection. The document is inserted with a new
	// unique ID, or the _id field is used if it is already set. This field is expected to be a string.
	// The result of the insert is returned.
	// This method returns a DuplicateKeyError if a document with the given ID already exists.
	// Any other rentalErrors are unexpected.
	Insert(ctx context.Context, collection string, document any) (ID, error)

	// FindOne returns a single document from the specified collection that matches the given filter
	// with the given options applied. If options is nil, default behaviour is used.
	// If no document is found, a NoDocumentsError is returned.
	FindOne(ctx context.Context, collection string, filter Filter, options *Options, result any) error

	// FindMany returns a slice of documents in the supplied type from the specified collection
	// that matches the given filter with the given options applied. If options is nil, default behaviour is used.
	// results must be a pointer to a slice.
	FindMany(ctx context.Context, collection string, filter Filter, options *Options, results any) error

	// UpdateOne applies the supplied update query on the first document with the given filter.
	// If upsert is true, a new document is created based on the filter if no document matched the filter.
	// Returns a NoDocumentsError if no document matched the filter and upsert was set to false.
	// Returns a DuplicateKeyError if upsert was set to true and the update query would create a document with a
	// duplicate ID.
	UpdateOne(ctx context.Context, collection string, filter Filter, update Update, upsert bool) error

	// DropCollection drops a given collection. This is a destructive relation and should only be used for testing.
	DropCollection(ctx context.Context, collection string) error

	// Aggregate retrieves a slice of documents from the specified collection generated by the given pipeline.
	// results must be a pointer to a slice.
	Aggregate(ctx context.Context, collection string, pipeline Pipeline, results any) error
}

// Options defines optional parameters for some query methods. Each field may be nil to use default behaviour.
type Options struct {
	// Sort specifies an order to sort matched documents by. By default, the database’s internal order is used.
	Sort Sort
	// Projection specifies a subset of fields to return. By default, all fields are returned.
	Projection Projection
}

type ID = string

type connection struct {
	database *mongo.Database
	client   *mongo.Client
	factory  QueryFactory
}

// NewDbConnection creates a new connection to the database. You should defer a call to the CleanUpDatabase method
// on the returned IConnection object.
// Any rentalErrors are unexpected.
func NewDbConnection(config DatabaseConfig) (IConnection, error) {
	m := connection{factory: &MongoFactory{}}
	return &m, m.setupDatabase(config)
}

func toConnectionUri(config DatabaseConfig) string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s",
		config.GetMongoDbUser(),
		config.GetMongoDbPassword(),
		config.GetMongoDbHost(),
		config.GetMongoDbPort(),
		config.GetMongoDbDatabase(),
	)
}

func (m *connection) setupDatabase(config DatabaseConfig) error {
	opts := mongoOptions.Client()
	opts.ApplyURI(toConnectionUri(config))
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	// ensure connection was successful
	if err = m.client.Ping(ctx, nil); err != nil {
		return err
	}

	m.database = m.client.Database(config.GetMongoDbDatabase(), mongoOptions.Database())
	return nil
}

func (m *connection) CleanUpDatabase() error {
	return m.client.Disconnect(context.Background())
}

func (m *connection) GetFactory() QueryFactory {
	return m.factory
}

func (m *connection) Insert(ctx context.Context, collection string, document any) (ID, error) {
	res, err := m.database.Collection(collection).InsertOne(ctx, document)
	if mongo.IsDuplicateKeyError(err) {
		return "", DuplicateKeyError
	}
	if err != nil {
		return "", err
	}
	return res.InsertedID.(ID), nil
}

func (m *connection) FindOne(ctx context.Context, collection string, filter Filter, options *Options, result any) error {
	opts := mongoOptions.FindOne()
	if options != nil && options.Projection != nil {
		opts = opts.SetProjection(options.Projection.getProjection())
	}
	if options != nil && options.Sort != nil {
		opts = opts.SetSort(options.Sort.getSort())
	}
	res := m.database.Collection(collection).FindOne(ctx, filter.getFilter(), opts)
	err := res.Decode(result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return NoDocumentsError
	}
	return err
}

func (m *connection) FindMany(ctx context.Context, collection string, filter Filter, options *Options,
	results any) error {

	opts := mongoOptions.Find()
	if options != nil && options.Projection != nil {
		opts = opts.SetProjection(options.Projection.getProjection())
	}
	if options != nil && options.Sort != nil {
		opts = opts.SetSort(options.Sort.getSort())
	}
	res, err := m.database.Collection(collection).Find(ctx, filter.getFilter(), opts)
	if err != nil {
		return err
	}
	return res.All(ctx, results)
}

func (m *connection) UpdateOne(ctx context.Context, collection string, filter Filter, update Update,
	upsert bool) error {

	opts := mongoOptions.Update()
	opts.SetUpsert(upsert)

	res, err := m.database.Collection(collection).UpdateOne(ctx, filter.getFilter(), update.getUpdate(), opts)
	if upsert && mongo.IsDuplicateKeyError(err) {
		return DuplicateKeyError
	}
	if err != nil {
		return err
	}
	if !upsert && res.MatchedCount == 0 {
		return NoDocumentsError
	}
	return nil
}

func (m *connection) DropCollection(ctx context.Context, collection string) error {
	return m.database.Collection(collection).Drop(ctx)
}

func (m *connection) Aggregate(ctx context.Context, collection string, pipeline Pipeline, results any) error {
	res, err := m.database.Collection(collection).Aggregate(ctx, pipeline.getPipeline())
	if err != nil {
		return err
	}
	return res.All(ctx, results)
}
