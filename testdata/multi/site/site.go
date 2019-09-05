package site

import (
	"example/site/comment"
	"example/site/user"
)

// Database is the master database that provides access to all databases.
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
