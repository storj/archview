package comment

import "example/site/user"

// Endpoint implements the comments repository.
//
// architecture: Database
type Repo interface {
}

// Endpoint implements the comments service.
//
// architecture: Service
type Service struct {
	comments Repo
}

// Endpoint implements the comments endpoint.
//
// architecture: Endpoint
type Endpoint struct {
	users    user.Repo
	comments *Service
}
