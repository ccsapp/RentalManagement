package db

import "errors"

var (
	DuplicateKeyError = errors.New("key does already exist")
	NoDocumentsError  = errors.New("no document matched the query")
)
