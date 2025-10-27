package svc

import (
	"fmt"
	"time"
	"watch-progress-service/services/gateway/api/internal/types"

	"github.com/gocql/gocql"
)

type Scylla struct {
	Db *gocql.Session
}

func (s *Scylla) getUserEpisodeUpdateQuery() string {
	return fmt.Sprintf("UPDATE %s USING TIMESTAMP ? SET last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ? AND episode_id = ?", episodeProgressTable)
}

func (s *Scylla) getUserMovieUpdateQuery() string {
	return fmt.Sprintf("UPDATE %s USING TIMESTAMP ? SET last_episode_id = ?, last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ?", movieProgressTable)
}

func (s *Scylla) SetLastWatchTime(data *types.SetWatchTimeReq) error {
	batch := s.Db.Batch(gocql.UnloggedBatch)
	episodeInserter := s.getUserEpisodeUpdateQuery()
	movieInserter := s.getUserMovieUpdateQuery()
	t, err := time.Parse(time.RFC3339, data.LastWatchedAt)
	if err != nil {
		return err
	}
	batch.Query(episodeInserter, t.UnixMicro(), data.EndTime, t, data.UserId, data.MovieId, data.EpisodeId)
	batch.Query(movieInserter, t.UnixMicro(), data.EpisodeId, data.EndTime, t, data.UserId, data.MovieId)
	if err := s.Db.ExecuteBatch(batch); err != nil {
		return err
	}
	return nil
}

func NewScylla(scyllaConn *gocql.Session) *Scylla {
	return &Scylla{
		scyllaConn,
	}
}
