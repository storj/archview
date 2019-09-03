// want package: "\\Qexample/site/comment.Repo[repo] = {}; example/site/comment.Service[service] = {comments:example/site/comment.Repo}; example/site/comment.Endpoint[endpoint] = {users:example/site/user.Repo, comments:example/site/comment.Service}\\E"
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
