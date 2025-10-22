package scylla

import (
	"github.com/gocql/gocql"
)

type Params struct {
	Clusters    []string
	Keyspace    string
	Username    string
	Password    string
	Consistency gocql.Consistency
}

func NewScyllaConn(cfg Params) (*gocql.Session, error) {
	var cluster = gocql.NewCluster(cfg.Clusters...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	cluster.Consistency = cfg.Consistency
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}
