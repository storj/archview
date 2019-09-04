// want package: "\\Qexample/pgdb.DB[database] = {}\\E"

package pgdb

import (
	"example/site/comment"
	"example/site/user"
)

// archview: database
type DB struct{}

func New() *DB {
	return &DB{}
}

func (db *DB) Users() user.Repo       { return db }
func (db *DB) Comments() comment.Repo { return db }
