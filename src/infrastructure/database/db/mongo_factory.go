package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func (f *MongoFactory) ArrayFilterAggregation(arrayName string, filter Filter, limit int, sort Sort) Pipeline {
	// Create an output document ("unwound document") for each array element of an input document.
	// Each output document is the input document with the value of the array field replaced by the element.
	unwindStage := bson.D{{"$unwind", "$" + arrayName}}

	// apply the filter to each unwound document
	matchStage := bson.D{{"$match", filter.getFilter()}}

	pipelineBuild := mongo.Pipeline{unwindStage, matchStage}

	if sort != nil {
		// Sort the unwound documents to bring the array elements of
		// a single input document in order.
		// This also orders the unwound documents of different input documents
		// with respect to each other. This ordering is later destroyed by $group,
		// but we need it here to apply the limit.
		sortStage := bson.D{{"$sort", sort.getSort()}}

		pipelineBuild = append(pipelineBuild, sortStage)
	}

	if limit > 0 {
		// limit the number of unwound documents to the total number of
		// array elements across all documents after processing
		limitStage := bson.D{{"$limit", limit}}

		pipelineBuild = append(pipelineBuild, limitStage)
	}

	// To "rewind" the unwound documents, we first group them by ID, rebuild
	// the array in the correct order, and place the rest of the document in the
	// "doc" field. At this point, "doc" also contains the first array element of
	// the array of the respective document. You should also note that $group destroys
	// the ordering of the output documents (but not the new order of the array elements).
	groupStage := bson.D{{"$group", bson.D{
		{"_id", "$_id"},
		{"doc", bson.D{{"$first", "$$ROOT"}}},
		{arrayName, bson.D{{"$push", "$" + arrayName}}}}}}

	// Merge "doc" and the array field into a new document. The array is merged into "doc"
	// to overwrite the array field of "doc" with the whole array (see above).
	replaceStage := bson.D{{"$replaceRoot",
		bson.D{{"newRoot",
			bson.D{{"$mergeObjects",
				bson.A{"$doc", bson.D{{arrayName, "$" + arrayName}}}}}}}}}

	pipelineBuild = append(pipelineBuild, groupStage, replaceStage)

	if sort != nil {
		// Sort documents a second time after modifying the array
		// to restore the correct order of the output documents.
		sortStage := bson.D{{"$sort", sort.getSort()}}

		pipelineBuild = append(pipelineBuild, sortStage)
	}

	return &pipeline{pipelineBuild}
}
