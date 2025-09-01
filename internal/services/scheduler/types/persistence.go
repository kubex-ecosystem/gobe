package types

// Database is an interface for database operations.
type Database interface {
	Query(query string, args ...interface{}) (Rows, error)
}

// Rows is an interface for iterating over database query results.
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}
