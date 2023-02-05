package db

import "go.mongodb.org/mongo-driver/bson"

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

func (f *MongoFactory) FilterLess(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$lt", value}}}}}
}

func (f *MongoFactory) FilterGreaterEqual(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$gte", value}}}}}
}

func (f *MongoFactory) FilterGreater(fieldName string, value interface{}) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$gt", value}}}}}
}

func (f *MongoFactory) FilterAnd(filter1 Filter, filter2 Filter) Filter {
	return &filter{bson.D{{"$and", bson.A{filter1.getFilter(), filter2.getFilter()}}}}
}

func (f *MongoFactory) FilterOr(filter1 Filter, filter2 Filter) Filter {
	return &filter{bson.D{{"$or", bson.A{filter1.getFilter(), filter2.getFilter()}}}}
}

func (f *MongoFactory) FilterNot(filterParam Filter) Filter {
	return &filter{bson.D{{"$nor", bson.A{filterParam.getFilter()}}}}
}

func (f *MongoFactory) FilterEverything() Filter {
	return &filter{bson.D{}}
}

func (f *MongoFactory) FilterElementMatch(fieldName string, filterParam Filter) Filter {
	return &filter{bson.D{{fieldName, bson.D{{"$elemMatch", filterParam.getFilter()}}}}}
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

func (f *MongoFactory) UpdatePush(fieldName string, value interface{}) Update {
	return &update{bson.D{{"$push", bson.D{{fieldName, value}}}}}
}
