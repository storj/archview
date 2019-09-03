// want package: "\\Qexample/site/user.Repo[database] = {}, example/site/user.Service[service] = {}, example/site/user.Endpoint[endpoint] = {}\\E"
package user

type Repo interface{} // archview: database

type Service struct {
	users Repo
} // archview: service

type Endpoint struct {
	users *Service
} // archview: endpoint
