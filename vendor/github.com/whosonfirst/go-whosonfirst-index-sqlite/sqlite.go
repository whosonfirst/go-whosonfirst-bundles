package sqlite

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/utils"
	"runtime"
	"strings"
)

func init() {
	dr := NewSQLiteDriver()
	index.Register("sqlite", dr)
}

type SQLiteDriver struct {
	index.Driver
}

func NewSQLiteDriver() index.Driver {
	return &SQLiteDriver{}
}

func (d *SQLiteDriver) Open(uri string) error {
	// check SQLite PRAGMAs here
	return nil
}

func (d *SQLiteDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	db, err := database.NewDB(uri)

	if err != nil {
		return err
	}

	defer db.Close()

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	has_table, err := utils.HasTable(db, "geojson")

	if err != nil {
		return err
	}

	if !has_table {
		return errors.New("database is missing a geojson table")
	}

	rows, err := conn.Query("SELECT id, body FROM geojson")

	if err != nil {
		return err
	}

	// https://github.com/whosonfirst/go-whosonfirst-index/issues/5

	sqlite_ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cpus := runtime.NumCPU() * 100 // configurable? (20171222/thisisaaronland)
	throttle_ch := make(chan bool, cpus)

	for i := 0; i < cpus; i++ {
		throttle_ch <- true
	}

	error_ch := make(chan error)

	for rows.Next() {

		<-throttle_ch

		var wofid int64
		var body string

		err := rows.Scan(&wofid, &body)

		if err != nil {
			return err
		}

		go func(ctx context.Context, wofid int64, body string, throttle_ch chan bool, error_ch chan error) {

			defer func() {
				throttle_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:

				// uri := fmt.Sprintf("sqlite://%s#geojson:%d", path, wofid)

				// see the way we're passing in STDIN and not uri as the path?
				// that because we call ctx, err := ContextForPath(path) in the
				// process() method and since uri won't be there nothing will
				// get indexed - it's not ideal it's just what it is today...
				// (20171213/thisisaaronland)

				fh := strings.NewReader(body)

				ctx = index.AssignPathContext(ctx, index.STDIN)
				err := index_cb(ctx, fh)

				if err != nil {
					error_ch <- err
				}
			}

		}(sqlite_ctx, wofid, body, throttle_ch, error_ch)

		select {
		case e := <-error_ch:
			cancel()
			return e
		default:
			// pass
		}
	}

	err = rows.Err()

	if err != nil {
		return err
	}

	return nil
}
