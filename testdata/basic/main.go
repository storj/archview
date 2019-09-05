package main

// architecture: Database
type DB interface{}

// architecture: Users
type Users struct {
	db       *DB
	comments *Comments
}

// architecture: Comments
type Comments struct {
	db    *DB
	users *Users
}

// architecture: Server
type Server struct {
	db       *DB
	comments *Comments
	users    *Users
}

func main() {}
