package site

import (
	"example/site/comment"
	"example/site/user"
)

type DB interface {
	Users() user.Repo
	Comments() comment.Repo
} // archview: Database

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
} // archview: Peer
