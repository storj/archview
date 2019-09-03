// want package: "\\Qexample/site.DB[database] = {Comments:example/site/comment.Repo, Users:example/site/user.Repo}; example/site.Site[peer] = {User.Endpoint:example/site/user.Endpoint, User.Service:example/site/user.Service, Comment.Endpoint:example/site/comment.Endpoint, Comment.Service:example/site/comment.Service}\\E"
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
