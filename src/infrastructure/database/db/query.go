package db

import "go.mongodb.org/mongo-driver/bson"

// QueryFactory defines methods that provide parameters to interact with a database
// through an instance of the IConnection interface.
type QueryFactory interface {

	// Create filter query parameters that is filters that select entries of the database

	// FilterMatch creates a filter that matches documents with the fields provided in the given bson annotated document
	FilterMatch(document interface{}) Filter
	// FilterEqual creates a filter that matches documents with the given field equal to the given value
	FilterEqual(fieldName string, value interface{}) Filter
	// FilterLessEqual creates a filter that matches documents with the given field less or equal to the given value
	FilterLessEqual(fieldName string, value interface{}) Filter
	// FilterGreaterEqual creates a filter that matches documents with the given field greater or equal to the given value
	FilterGreaterEqual(fieldName string, value interface{}) Filter
	// FilterAnd creates a filter that matches documents that matches filter1 and filter2
	FilterAnd(filter1 Filter, filter2 Filter) Filter
	// FilterOr creates a filter that matches documents that matches filter1 or filter2
	FilterOr(filter1 Filter, filter2 Filter) Filter
	// FilterEverything creates a filter that matches any document
	FilterEverything() Filter

	// Create sort query parameters that is a sorting order for the returned documents

	// SortAsc creates a sorting order that sorts documents in ascending order based on field
	SortAsc(fieldName string) Sort
	// SortDesc creates a sorting order that sorts documents in descending order based on field
	SortDesc(fieldName string) Sort

	// Create projection query parameters that is a selection of fields from the documents

	// ProjectionSingle creates a projection that selects the given field
	ProjectionSingle(fieldName string) Projection
	// ProjectionID creates a projection that selects the id
	ProjectionID() Projection

	// Create update query parameters that is a request to change documents in the database

	// UpdateSingle creates an update request that sets field to value. You cannot change a document’s id.
	UpdateSingle(fieldName string, value interface{}) Update
	// UpdateMultiple creates an update request that updates all fields present in document.
	// The IDs of the documents must match if supplied.
	UpdateMultiple(document interface{}) Update
}

type Filter interface {
	getFilter() any
}

type filter struct {
	filter any
}

func (f *filter) getFilter() any {
	return f.filter
}

type Sort interface {
	getSort() any
}

type sort struct {
	sort any
}

func (s *sort) getSort() any {
	return s.sort
}

type Projection interface {
	getProjection() any
}

type projection struct {
	projection any
}

func (p *projection) getProjection() any {
	return p.projection
}

type Update interface {
	getUpdate() any
}

type update struct {
	update any
}

func (u *update) getUpdate() any {
	return u.update
}

type MongoFactory struct{}

func (f *MongoFactory) FilterMatch(document interface{}) Filter {
	return &filter{document}
}

func (f *MongoFactory) FilterEqual(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, value}}}
}

func (f *MongoFactory) FilterLessEqual(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$lte", value}}}}}
}

func (f *MongoFactory) FilterGreaterEqual(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$gte", value}}}}}
}

func (f *MongoFactory) FilterAnd(filter1 Filter, filter2 Filter) Filter {
	return &filter{bson.D{{"$and", bson.A{filter1, filter2}}}}
}

func (f *MongoFactory) FilterOr(filter1 Filter, filter2 Filter) Filter {
	return &filter{bson.D{{"$or", bson.A{filter1, filter2}}}}
}

func (f *MongoFactory) FilterEverything() Filter {
	return &filter{bson.D{}}
}

func (f *MongoFactory) SortAsc(fieldName string) Sort {
	return &sort{bson.D{{fieldName, 1}}}
}

func (f *MongoFactory) SortDesc(fieldName string) Sort {
	return &sort{bson.D{{fieldName, -1}}}
}

func (f *MongoFactory) ProjectionSingle(fieldName string) Projection {
	if fieldName == "_id" {
		return f.ProjectionID()
	}
	return &projection{bson.D{{"_id", 0}, {fieldName, 1}}}
}

func (f *MongoFactory) ProjectionID() Projection {
	return &projection{bson.D{{"_id", 1}}}
}

func (f *MongoFactory) UpdateSingle(fieldName string, value interface{}) Update {
	return &update{bson.D{{"$set", bson.D{{fieldName, value}}}}}
}

func (f *MongoFactory) UpdateMultiple(document interface{}) Update {
	return &update{bson.D{{"$set", document}}}
}
