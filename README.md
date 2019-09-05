# ArchView

ArchView is a tool for creating and inspecting Go program architecture based on annotations.


```
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
	db       *DB
	comments *Comments
	users    *Users
}
```

Allows creating a diagram that looks like:

![Basic Diagram](testdata/basic/graph.svg)