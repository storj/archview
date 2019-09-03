// want package: "\\Qexample/site.DB[database] = {}, example/site.Site[peer] = {}\\E"
package site

import (
	"example/site/comment"
	"example/site/user"
)

type DB interface {
	Users() user.Repo
	Comments() comment.Repo
} // archview: database

type Site struct {
	User struct {
		Endpoint *user.Endpoint
		Service  *user.Service
	}

	Comment struct {
		Endpoint *comment.Endpoint
		Service  *comment.Service
	}
} // archview: peer
