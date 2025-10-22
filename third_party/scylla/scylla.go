package scylla

import (
	"github.com/gocql/gocql"
)

type Params struct {
	Host        string
	Port        uint16
	Keyspace    string
	Username    string
	Password    string
	Consistency uint16
}

type Client struct {
	Db  *gocql.Session
	cfg Params
}

func NewClient(cfg Params) (*Client, error) {
	var cluster = gocql.NewCluster("localhost:9042")
	cluster.Keyspace = "belet"
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Consistency = gocql.One
	var session, err = cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &Client{
		session,
		cfg,
	}, nil
}
