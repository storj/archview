// want package: "\\Qexample/site/user.Repo[database] = {}; example/site/user.Service[service] = {users:example/site/user.Repo}; example/site/user.Endpoint[endpoint] = {users:example/site/user.Service}\\E"
package user

type Repo interface{} // archview: database

type Service struct {
	users Repo
} // archview: service

type Endpoint struct {
	users *Service
} // archview: endpoint
