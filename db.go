package rest

import "database/sql"

type DbOpener interface {
	// Open creates a database connection against databaseName
	// If databaseName is empty, uses Connection URL from database broker
	Open(databaseName string) (DbBroker, error)
}

type DbBroker interface {
	ConnectionUrl() string
	Db() *sql.DB
	Close() error
}
