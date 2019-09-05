package main

// DB is database.
//
// architecture: Database
type DB interface{}

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
