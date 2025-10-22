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

type Client struct {
	Db  *gocql.Session
	cfg Params
}

func NewClient(cfg Params) (*Client, error) {
	var cluster = gocql.NewCluster(cfg.Clusters...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	cluster.Consistency = cfg.Consistency
	var session, err = cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &Client{
		session,
		cfg,
	}, nil
}
