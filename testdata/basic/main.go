package main

// DB is database.
//
// architecture: Database
type DB interface {
	Query() string
}

// Users is a service.
//
// architecture: Service
type Users struct {
	db       *DB
	comments *Comments
}

// Comments is a service.
//
// architecture: Service
type Comments struct {
	db    *DB
	users *Users
}

// Server is a server.
//
// architecture: Server
type Server struct {
	comments *Comments
	users    *Users

	node *node
}

// test for recursion
type node struct {
	next *node
}

var _ DB = &PostgresDB{}

// PostgresDB implements DB
//
// architecture: Database Implementation
type PostgresDB struct{}

// Query returns a string.
func (db *PostgresDB) Query() string { return "string" }

var _ DB = (*SqliteDB)(nil)

// SqliteDB implements DB
//
// architecture: Database Implementation
type SqliteDB struct{}

// Query returns a string.
func (db *SqliteDB) Query() string { return "string" }
