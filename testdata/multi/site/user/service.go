package user

// Repo implements the users repo.
//
// architecture: Database
type Repo interface {
}

// Service implements the users service.
//
// architecture: Service
type Service struct {
	users Repo
}

// Endpoint implements the users endpoint.
//
// architecture: Endpoint
type Endpoint struct {
	users *Service
}
