package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"testing"
	"time"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/services/gateway/api/internal/handler"
	"watch-progress-service/services/gateway/api/internal/svc"
	"watch-progress-service/services/gateway/api/internal/types"
	"watch-progress-service/third_party/scylla"

	"github.com/testcontainers/testcontainers-go/modules/scylladb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

const episodeProgressTable = "user_episode_progress"
const movieProgressTable = "user_movie_progress"

type IntegrationTestSuite struct {
	suite.Suite
	scyllaContainer *scylladb.Container
	scyllaSession   *gocql.Session
	server          *rest.Server
	testConfig      config.Config
}

func waitForServer(url string) error {
	client := &http.Client{Timeout: 1 * time.Second}
	for attempt := 1; attempt <= 20; attempt++ { // Up to ~20s
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logx.Info("Test server ready")
				return nil
			}
		}
		logx.Infof("Waiting for server (attempt %d): %v", attempt, err)
		time.Sleep(1 * time.Second)
	}
	return errors.New("test server did not become ready in time")
}

func (s *IntegrationTestSuite) setupScyllaDb() {
	err := s.scyllaSession.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, s.testConfig.Scylla.Keyspace)).Exec()
	s.Require().NoError(err)

	err = s.scyllaSession.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		user_id int,
		movie_id int,
		episode_id int,
		last_watch_time float,
		last_watched_at timestamp,
		PRIMARY KEY ((user_id, movie_id), episode_id)
	) WITH CLUSTERING ORDER BY (episode_id ASC);`, s.testConfig.Scylla.Keyspace, episodeProgressTable)).Exec()
	s.Require().NoError(err)

	err = s.scyllaSession.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		user_id int,
		movie_id int,
		last_episode_id int,
		last_watch_time float,
		last_watched_at timestamp,
		PRIMARY KEY (user_id, movie_id)
	) WITH CLUSTERING ORDER BY (movie_id ASC);`, s.testConfig.Scylla.Keyspace, movieProgressTable)).Exec()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error

	s.testConfig = config.Config{
		Rest: rest.RestConf{
			Host: "127.0.0.1",
			Port: 9999,
		},
		Scylla: scylla.Params{
			Clusters: []string{"127.0.0.1:9042"},
			Keyspace: "belet",
			Username: "cassandra",
			Password: "cassandra",
		},
		ApiKeys: []string{"test_api_key"},
	}

	initialCluster := gocql.NewCluster(s.testConfig.Scylla.Clusters...)
	initialCluster.Consistency = gocql.LocalQuorum
	initialCluster.ConnectTimeout = 10 * time.Second
	initialCluster.Timeout = 10 * time.Second

	// Retry connection for initial session (containers can take time)
	var initialSession *gocql.Session
	err = retry.Do(
		func() error {
			initialSession, err = initialCluster.CreateSession()
			return err
		},
		retry.Attempts(10),
		retry.Delay(2*time.Second),
	)
	s.Require().NoError(err)
	defer initialSession.Close() // Close after keyspace creation

	// Create keyspace if not exists
	createKeyspaceQuery := `CREATE KEYSPACE IF NOT EXISTS belet 
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`
	err = initialSession.Query(createKeyspaceQuery).Exec()
	s.Require().NoError(err)

	cluster := gocql.NewCluster(s.testConfig.Scylla.Clusters...)
	cluster.Keyspace = s.testConfig.Scylla.Keyspace
	cluster.Consistency = gocql.LocalQuorum
	s.scyllaSession, err = cluster.CreateSession()
	s.Require().NoError(err)

	s.setupScyllaDb()
	s.Require().NoError(err)

	svcCtx, err := svc.NewServiceContext(s.testConfig)
	s.Require().NoError(err)

	s.server = rest.MustNewServer(s.testConfig.Rest)
	handler.RegisterHandlers(s.server, svcCtx)

	s.server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		},
	})

	// Start the server in background with error handling
	startErrChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				startErrChan <- fmt.Errorf("%v", r)
			}
		}()

		logx.Info("Starting test server...")
		s.server.Start()
		close(startErrChan)
	}()

	// Wait for startup error (quick check)
	select {
	case err := <-startErrChan:
		if err != nil {
			s.Require().NoError(err) // Will fail suite with error message
		}
	case <-time.After(1 * time.Second): // Give it a moment; no immediate error
	}

	// Wait for server readiness (updated helper below)
	err = waitForServer(fmt.Sprintf("http://127.0.0.1:%d/health", s.testConfig.Rest.Port))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	_ = s.scyllaSession.Query(fmt.Sprintf(`DROP KEYSPACE IF EXISTS %s`, s.testConfig.Scylla.Keyspace)).Exec()

	if s.scyllaSession != nil {
		s.scyllaSession.Close()
	}
	if s.scyllaContainer != nil {
		err := s.scyllaContainer.Terminate(context.Background())
		s.Require().NoError(err)
	}
	if s.server != nil {
		s.server.Stop()
	}
}

func (s *IntegrationTestSuite) TestSetLastWatchTimeSuccess() {
	reqBody := types.SetWatchTimeReq{
		UserId:        1,
		MovieId:       1,
		EpisodeId:     1,
		StartTime:     10,
		EndTime:       20,
		Duration:      7680,
		LastWatchedAt: "2025-10-28T15:28:36+05:00",
	}
	jsonData, err := json.Marshal(reqBody)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s:%d/api/v1/setLastWatchTime/", s.testConfig.Rest.Host, s.testConfig.Rest.Port)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test_api_key")

	client := &http.Client{}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	fmt.Printf("Response: %s", string(body))
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var episodeTableRowCount, movieTableRowCount int
	err = s.scyllaSession.Query(fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_id = ? and movie_id = ? and episode_id = ?`, episodeProgressTable),
		reqBody.UserId, reqBody.MovieId, reqBody.EpisodeId).Scan(&episodeTableRowCount)
	s.Require().NoError(err)

	err = s.scyllaSession.Query(fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_id = ? and movie_id = ?`, movieProgressTable),
		reqBody.UserId, reqBody.MovieId).Scan(&movieTableRowCount)
	s.Require().NoError(err)

	s.Equal(1, episodeTableRowCount)
	s.Equal(1, movieTableRowCount)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
