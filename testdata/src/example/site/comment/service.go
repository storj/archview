// want package: "\\Qexample/site/comment.Repo[repo] = {}, example/site/comment.Service[service] = {}, example/site/comment.Endpoint[endpoint] = {}\\E"
package comment

import "example/site/user"

type Repo interface{} // archview: repo

type Service struct {
	comments Repo
} // archview: service

type Endpoint struct {
	users    user.Repo
	comments *Service
} // archview: endpoint
