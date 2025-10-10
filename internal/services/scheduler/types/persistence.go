package types

import "context"

type DBService interface {
	Database() DirectDatabase
}

type DirectDatabase interface {
	Query(context.Context, string, ...interface{}) (any, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}
