package site

import (
	"example/site/comment"
	"example/site/user"
)

// DB is the master database that provides access to all databases.
// architecture: Database
type DB interface {
	Users() user.Repo
	Comments() comment.Repo
}

// Site implements the main implementation class.
// architecture: Peer
type Site struct {
	DB DB

	User struct {
		Endpoint *user.Endpoint
		Service  *user.Service
	}

	Comment struct {
		Endpoint *comment.Endpoint
		Service  *comment.Service
	}
}

// ServiceOnly implements a separate peer class.
// architecture: Peer
type ServiceOnly struct {
	DB DB

	User struct {
		Service  *user.Service
	}

	Comment struct {
		Service  *comment.Service
	}
}
