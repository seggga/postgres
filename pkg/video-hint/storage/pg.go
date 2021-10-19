package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type ContextKey int

const ContextKeyDB ContextKey = iota + 1

var (
	db    *pgxpool.Pool
	dbMux = &sync.Mutex{}
)

type FoundVideo struct {
	Caption  string `json:"caption"`
	URI      string `json:"uri"`
	Location string `json:"location"`
}

/*

try then

*/

type DB interface {
	GetVideosByCaption(ctx context.Context, prefix string) ([]*FoundVideo, error)
	Close()
}

type ConnString struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDB(connStr *ConnString) (DB, error) {
	// pool, err := getConn(connStr)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get a connection pool: %w", err)
	// }
	// return &conn{
	// 	db: pool,
	// }, nil
	gormDB, err := newGormDB(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open a gorm connection: %w", err)
	}
	return gormDB, nil
}

type conn struct {
	db *pgxpool.Pool
}

// GetVideosByCaption sends query to the DB and processes the given result
func (c *conn) GetVideosByCaption(ctx context.Context, phrase string) ([]*FoundVideo, error) {
	rows, err := c.db.Query(
		context.Background(),
		`SELECT caption, uri, location
		FROM videos
		WHERE caption LIKE '%' || $1 || '%'`,
		phrase,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	videos := make([]*FoundVideo, 0)
	for rows.Next() {
		v := &FoundVideo{}
		if err := rows.Scan(&v.Caption, &v.URI, &v.Location); err != nil {
			return nil, fmt.Errorf("failed to get rows on given phrase: %w", err)
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func (c *conn) Close() {
	c.db.Close()
}

func getConn(connStr *ConnString) (*pgxpool.Pool, error) {
	dbMux.Lock()
	defer dbMux.Unlock()
	if db != nil {
		return db, nil
	}

	var err error
	db, err = initPGXPool(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a PGX pool: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return db, nil
}

func initPGXPool(c *ConnString) (*pgxpool.Pool, error) {
	connStr, err := composeConnectionString(c)
	if err != nil {
		return nil, fmt.Errorf("failed to compose the connection string: %w", err)
	}
	cfg, err := getPGXPoolConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get the PGX pool config: %w", err)
	}
	db, err = pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres DB using a PGX connection pool: %w", err)
	}
	return db, nil
}

func getPGXPoolConfig(connStr string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create the PGX pool config from connection string: %w", err)
	}
	cfg.ConnConfig.ConnectTimeout = time.Second * 1
	cfg.ConnConfig.Logger = logrusadapter.NewLogger(
		&logrus.Logger{
			Out:          os.Stdout,
			Formatter:    new(logrus.JSONFormatter),
			Hooks:        make(logrus.LevelHooks),
			Level:        logrus.InfoLevel,
			ExitFunc:     os.Exit,
			ReportCaller: false,
		})
	return cfg, nil
}

func composeConnectionString(c *ConnString) (string, error) {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		url.QueryEscape(c.Host),
		url.QueryEscape(c.Port),
		url.QueryEscape(c.DBName),
	), nil
}
