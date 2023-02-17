package db

import (
	"errors"
	"strings"
)

// PseudoFactory is a factory for creating pseudo queries that do not actually do anything.
// They are used for testing purposes.
type PseudoFactory struct{}

type pseudoFilter struct {
	relation  string
	fieldName any
	value     any
}

type pseudoMultiFilter struct {
	relation string
	filter1  any
	filter2  any
}

type pseudoSort struct {
	order     any
	fieldName any
}

type pseudoProjection struct {
	fields string
}

type pseudoUpdate struct {
	field string
	value any
}

type pseudoNestedFilter struct {
	arrayName string
	filter    Filter
	limit     int
	sort      Sort
}

func (f *PseudoFactory) FilterMatch(document interface{}) Filter {
	return &filter{document}
}

func (f *PseudoFactory) FilterEqual(fieldName string, value interface{}) Filter {
	return &filter{pseudoFilter{"equal", fieldName, value}}
}

func (f *PseudoFactory) FilterLessEqual(fieldName string, value interface{}) Filter {
	return &filter{pseudoFilter{"lessEqual", fieldName, value}}
}

func (f *PseudoFactory) FilterLess(fieldName string, value interface{}) Filter {
	return &filter{pseudoFilter{"less", fieldName, value}}
}

func (f *PseudoFactory) FilterGreaterEqual(fieldName string, value interface{}) Filter {
	return &filter{pseudoFilter{"greaterEqual", fieldName, value}}
}

func (f *PseudoFactory) FilterGreater(fieldName string, value interface{}) Filter {
	return &filter{pseudoFilter{"greater", fieldName, value}}
}

func (f *PseudoFactory) FilterAnd(filter1 Filter, filter2 Filter) Filter {
	return &filter{pseudoMultiFilter{"and", filter1, filter2}}
}

func (f *PseudoFactory) FilterOr(filter1 Filter, filter2 Filter) Filter {
	return &filter{pseudoMultiFilter{"or", filter1, filter2}}
}

func (f *PseudoFactory) FilterNot(filterParam Filter) Filter {
	return &filter{pseudoMultiFilter{"not", filterParam, nil}}
}

func (f *PseudoFactory) FilterEverything() Filter {
	return &filter{pseudoFilter{"everything", "", ""}}
}

func (f *PseudoFactory) FilterElementMatch(fieldName string, filterParam Filter) Filter {
	return &filter{pseudoFilter{"elementMatch", fieldName, filterParam}}
}

func (f *PseudoFactory) SortAsc(fieldName string) Sort {
	return &sort{pseudoSort{"ascending", fieldName}}
}

func (f *PseudoFactory) SortDesc(fieldName string) Sort {
	return &sort{pseudoSort{"descending", fieldName}}
}

func (f *PseudoFactory) ProjectionSingle(fieldName string) Projection {
	return &projection{pseudoProjection{fieldName}}
}

func (f *PseudoFactory) ProjectionID() Projection {
	return &projection{pseudoProjection{"#+#+/ID/+#+#"}}
}

func (f *PseudoFactory) UpdateSingle(fieldName string, value interface{}) Update {
	return &update{pseudoUpdate{fieldName, value}}
}

func (f *PseudoFactory) UpdateMultiple(document interface{}) Update {
	return &update{pseudoUpdate{"#+#+/MULTIPLE/+#+#", document}}
}

func (f *PseudoFactory) UpdatePush(fieldName string, value interface{}) Update {
	return &update{pseudoUpdate{"#+#+/PUSH/+#+# TO " + fieldName, value}}
}

// UnpackPushUpdate unpacks a push update into the field name and the value to push.
// Returns an error if the update is not a push update created by this factory.
func (f *PseudoFactory) UnpackPushUpdate(updateParam Update) (string, interface{}, error) {
	pseudoUpd, ok := updateParam.getUpdate().(pseudoUpdate)
	if !ok {
		return "", nil, errors.New("updateParam is not a pseudo update")
	}

	if !strings.HasPrefix(pseudoUpd.field, "#+#+/PUSH/+#+# TO ") {
		return "", nil, errors.New("updateParam is not a pseudo push update")
	}

	return pseudoUpd.field[18:], pseudoUpd.value, nil
}

func (f *PseudoFactory) UpdateMatchingArrayElement(arrayName string, elementFieldName string,
	value interface{}) Update {

	return &update{pseudoUpdate{"#+#+/MATCHING/+#+# IN " + arrayName + ", SET " + elementFieldName, value}}
}

func (f *PseudoFactory) ArrayFilterAggregation(arrayName string, filter Filter, limit int, sort Sort) Pipeline {
	return &pipeline{pseudoNestedFilter{arrayName, filter, limit, sort}}
}
