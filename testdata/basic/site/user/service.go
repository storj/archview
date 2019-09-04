package user

type Repo interface{} // archview: Database

type Service struct {
	users Repo
} // archview: Service

type Endpoint struct {
	users *Service
} // archview: Endpoint
