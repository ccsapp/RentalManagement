package db

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
	// FilterLess creates a filter that matches documents with the given field less than the given value
	FilterLess(fieldName string, value interface{}) Filter
	// FilterGreaterEqual creates a filter that matches documents with the given field greater or equal to the given value
	FilterGreaterEqual(fieldName string, value interface{}) Filter
	// FilterGreater creates a filter that matches documents with the given field greater than the given value
	FilterGreater(fieldName string, value interface{}) Filter
	// FilterAnd creates a filter that matches documents that matches filter1 and filter2
	FilterAnd(filter1 Filter, filter2 Filter) Filter
	// FilterOr creates a filter that matches documents that matches filter1 or filter2
	FilterOr(filter1 Filter, filter2 Filter) Filter
	// FilterNot creates a filter that matches documents that do not match filter
	FilterNot(filter Filter) Filter
	// FilterEverything creates a filter that matches any document
	FilterEverything() Filter

	// FilterElementMatch creates a filter that matches documents that have an array field with at least one element
	// that matches the filter
	FilterElementMatch(fieldName string, filter Filter) Filter

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
	// UpdatePush creates an update request that adds value to an array field.
	UpdatePush(fieldName string, value interface{}) Update
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