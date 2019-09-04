package comment

import "example/site/user"

type Repo interface{} // archview: Database

type Service struct {
	comments Repo
} // archview: Service

type Endpoint struct {
	users    user.Repo
	comments *Service
} // archview: Endpoint
