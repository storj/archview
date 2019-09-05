package main

// architecture: Database
type DB interface{}

// architecture: Service
type Users struct {
	db       *DB
	comments *Comments
}

// architecture: Service
type Comments struct {
	db    *DB
	users *Users
}

// architecture: Server
type Server struct {
	comments *Comments
	users    *Users
}

func main() {}
