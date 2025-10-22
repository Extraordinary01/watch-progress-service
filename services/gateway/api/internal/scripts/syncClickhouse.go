package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gocql/gocql"
	"github.com/zeromicro/go-zero/core/conf"
	"log"
	"os"
	"time"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/third_party/clickhouse"
	"watch-progress-service/third_party/scylla"
)

var configFile = flag.String("f", os.Getenv("WATCH_PROGRESS_CONFIG_FILE"), "the config file")

type NextWatchData struct {
	userId        uint32    `db:"user_id"`
	movieId       uint32    `db:"movie_id"`
	originId      uint32    `db:"origin_id"`
	watchTime     float64   `db:"watch_time"`
	lastWatchedAt time.Time `db:"date"`
}

func syncData(ctx context.Context, scyllaConn *gocql.Session, clickhouseConn driver.Conn) (int, error) {
	usersCount := 120000
	selectNextToWatchQuery := "select user_id, movie_id, origin_id, watch_time, date from default.next_to_watch_mv where is_deleted = false and finished_watchin = false and user_id >= ? and user_id <= ?"
	updateEpisodeQuery := "UPDATE user_episode_progress USING TIMESTAMP ? SET last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ? AND episode_id = ?"
	updateMovieQuery := "UPDATE user_movie_progress USING TIMESTAMP ? SET last_episode_id = ?, last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ?"
	for i := 1; i <= usersCount; i += 1000 {
		rows, err := clickhouseConn.Query(ctx, selectNextToWatchQuery, i, i+999)
		if err != nil {
			return i, err
		}
		for rows.Next() {
			var data NextWatchData
			if err := rows.Scan(&data); err != nil {
				return i, err
			}
			batch := scyllaConn.Batch(gocql.LoggedBatch)
			batch.Query(updateEpisodeQuery, data.lastWatchedAt.UnixMicro(), data.watchTime, data.lastWatchedAt, data.userId, data.originId, data.movieId)
			batch.Query(updateMovieQuery, data.lastWatchedAt.UnixMicro(), data.movieId, data.watchTime, data.lastWatchedAt, data.userId, data.originId)
			if err := batch.Exec(); err != nil {
				return i, err
			}
		}
	}
	return usersCount, nil
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	scyllaConn, err := scylla.NewScyllaConn(c.Scylla)
	if err != nil {
		log.Fatalf("Failed to create Scylla connection, error: %s", err)
	}
	defer scyllaConn.Close()

	ctx := context.Background()

	clickhouseConn, err := clickhouse.NewClickhouseConn(ctx, c.Clickhouse)
	if err != nil {
		log.Fatalf("Failed to create ClickHouse connection, error: %s", err)
	}
	defer clickhouseConn.Close()

	scannedCount, err := syncData(ctx, scyllaConn, clickhouseConn)
	if err != nil {
		log.Fatalf("Failed to sync data, error: %s", err)
	}
	fmt.Printf("Synced %d records\n", scannedCount)
}
