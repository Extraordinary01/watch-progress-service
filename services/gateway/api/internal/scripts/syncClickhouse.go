package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gocql/gocql"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"os"
	"sync"
	"time"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/third_party/clickhouse"
	"watch-progress-service/third_party/scylla"
)

var configFile = flag.String("f", os.Getenv("WATCH_PROGRESS_CONFIG_FILE"), "the config file")

type NextWatchData struct {
	userId        uint32    `ch:"user_id"`
	movieId       uint32    `ch:"movie_id"`
	originId      uint32    `ch:"origin_id"`
	watchTime     float32   `ch:"watch_time"`
	lastWatchedAt time.Time `ch:"date"`
}

type Params struct {
	countOfUsers *int
	rangeOfUsers *int
	workerCount  *int
	startPoint   *int
	minDate      time.Time
}

func processClickhouseRows(rows driver.Rows, scyllaConn *gocql.Session, workerCount int) error {
	// Step 1: Read all rows
	var allData []NextWatchData
	updateEpisodeQuery := "UPDATE user_episode_progress USING TIMESTAMP ? SET last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ? AND episode_id = ?"
	updateMovieQuery := "UPDATE user_movie_progress USING TIMESTAMP ? SET last_episode_id = ?, last_watch_time = ?, last_watched_at = ? WHERE user_id = ? AND movie_id = ?"
	for rows.Next() {
		var d NextWatchData
		if err := rows.Scan(&d.userId, &d.movieId, &d.originId, &d.watchTime, &d.lastWatchedAt); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}
		allData = append(allData, d)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	// Step 2: Split into chunks
	chunkSize := (len(allData) + workerCount - 1) / workerCount
	errCh := make(chan error, workerCount)
	var wg sync.WaitGroup

	// Step 3: Process in parallel
	for i := 0; i < workerCount; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(allData) {
			end = len(allData)
		}
		if start >= len(allData) {
			break
		}

		wg.Add(1)
		go func(part []NextWatchData) {
			defer wg.Done()
			for _, data := range part {
				batch := scyllaConn.Batch(gocql.LoggedBatch)
				batch.Query(updateEpisodeQuery,
					data.lastWatchedAt.UnixMicro(),
					data.watchTime,
					data.lastWatchedAt,
					data.userId,
					data.originId,
					data.movieId,
				)
				batch.Query(updateMovieQuery,
					data.lastWatchedAt.UnixMicro(),
					data.movieId,
					data.watchTime,
					data.lastWatchedAt,
					data.userId,
					data.originId,
				)
				if err := scyllaConn.ExecuteBatch(batch); err != nil {
					errCh <- fmt.Errorf("batch update failed (user_id=%d): %w", data.userId, err)
					return
				}
			}
		}(allData[start:end])
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}
	return nil
}

func syncData(ctx context.Context, scyllaConn *gocql.Session, clickhouseConn driver.Conn, params *Params) (int, error) {
	selectNextToWatchQuery := "select user_id, movie_id, origin_id, cast(watch_time as Float32), date from default.next_to_watch_mv where is_deleted = false and finished_watching = false and user_id >= ? and user_id <= ? and date >= ?"
	for i := *params.startPoint; i <= *params.countOfUsers; i += *params.rangeOfUsers {
		rows, err := clickhouseConn.Query(ctx, selectNextToWatchQuery, i, i+*params.rangeOfUsers-1, params.minDate)
		if err != nil {
			return i, err
		}
		defer rows.Close()
		logx.Info(fmt.Sprintf("fetched next to watch data of users %d - %d", i, i+9999))
		if err := processClickhouseRows(rows, scyllaConn, *params.workerCount); err != nil {
			return i, err
		}
	}
	return *params.countOfUsers, nil
}

func main() {
	countOfUsers := flag.Int("count", 1633000, "number of users")
	rangeOfSelect := flag.Int("range", 10000, "range of user selection")
	workers := flag.Int("workers", 60, "number of workers (goroutines)")
	startPoint := flag.Int("start", 1, "id of user where selection will start")
	rawMinDate := flag.String("min_date", "2025-10-23 18:35:00", "min date for filtering data selection")

	minDate, err := time.Parse("2006-01-02 15:04:05", *rawMinDate)
	if err != nil {
		log.Fatalf("failed to parse min date: %v", err)
	}

	flag.Parse()
	params := Params{countOfUsers: countOfUsers, rangeOfUsers: rangeOfSelect, workerCount: workers,
		startPoint: startPoint, minDate: minDate}

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

	logx.Info("Starting data synchronization")
	scannedCount, err := syncData(ctx, scyllaConn, clickhouseConn, &params)
	if err != nil {
		log.Fatalf("Failed to sync data, error: %s", err)
	}
	fmt.Printf("Synced %d records\n", scannedCount)
}
