//go:build integration_tests
// +build integration_tests

package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/seggga/postgres/pkg/videos-hint/storage"
)

const (
	DB_HOST     = "127.0.0.1"
	DB_USER     = "gotuber"
	DB_PASSWORD = "Passw0rd"
	DB_NAME     = "go_tube"
)

var DB_PORT = ""

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	setupResult, err := setup()
	if err != nil {
		log.Println("setup err: ", err)
		return -1
	}
	defer teardown(setupResult)
	return m.Run()
}

type teardownPack struct {
	OldEnvVars map[string]string
}

const dataDir = "data"

type setupResult struct {
	Pool              *dockertest.Pool
	PostgresContainer *dockertest.Resource
}

const dockerMaxWait = time.Second * 5

func setup() (r *setupResult, err error) {
	testFileDir, err := getTestFileDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get the script dir: %w", err)
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docketest pool: %w", err)
	}
	pool.MaxWait = dockerMaxWait

	postgresContainer, err := runPostgresContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the Postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			if err := pool.Purge(postgresContainer); err != nil {
				log.Println("failed to purge the postgres container: %w", err)
			}
		}
	}()

	migrationContainer, err := runMigrationContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the migration container: %w", err)
	}

	defer func() {
		if err := pool.Purge(migrationContainer); err != nil {
			err = fmt.Errorf("failed to purge the migration container: %w", err)
		}
	}()

	if err := pool.Retry(func() error {
		err := prepopulateDB(testFileDir)
		if err != nil {
			log.Printf("populate DB err: %v", err)
		}
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to prepopulate the DB: %w", err)
	}

	return &setupResult{
		Pool:              pool,
		PostgresContainer: postgresContainer,
	}, nil
}

func getTestFileDir() (string, error) {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get the caller info")
	}
	fileDir := filepath.Dir(fileName)
	dir, err := filepath.Abs(fileDir)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path to the directory %s: %w", dir, err)
	}
	return fileDir, nil
}

func runPostgresContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	postgresContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "14.0",
			Env: []string{
				"POSTGRES_PASSWORD=P@ssw0rd",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/docker-entrypoint-initdb.d",
					Source: filepath.Join(testFileDir, "init"),
					Type:   "bind",
				},
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the postgres docker container: %w", err)
	}
	postgresContainer.Expire(120)

	DB_PORT = postgresContainer.GetPort("5432/tcp")

	// Wait for the DB to start
	if err := pool.Retry(func() error {
		db, err := getDBConnector()
		if err != nil {
			return fmt.Errorf("failed to get a DB connector: %w", err)
		}
		return db.Ping(context.Background())
	}); err != nil {
		pool.Purge(postgresContainer)
		return nil, fmt.Errorf("failed to ping the created DB: %w", err)
	}
	return postgresContainer, nil
}

func runMigrationContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	migrationsDir, err := filepath.Abs(filepath.Join(testFileDir, "../../../../migrations"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path of the migrations dir: %w", err)
	}
	migrationContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "migrate/migrate",
			Tag:        "v4.15.0",
			Cmd: []string{
				"-path=/migrations",
				fmt.Sprintf(
					"-database=%s",
					composeConnectionString(),
				),
				"up",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/migrations",
					Source: migrationsDir,
					Type:   "bind",
				},
			}
			config.NetworkMode = "host"
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the migration container: %w", err)
	}

	return migrationContainer, err
}

func prepopulateDB(testFileDir string) error {
	prepopulateScriptPath := filepath.Join(testFileDir, "prepopulate_db.sql")
	scriptBytes, err := os.ReadFile(prepopulateScriptPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", prepopulateScriptPath, err)
	}
	conn, err := getDBConnector()
	if err != nil {
		return fmt.Errorf("failed to get a DB connector: %w", err)
	}
	if _, err := conn.Exec(context.Background(), string(scriptBytes)); err != nil {
		return fmt.Errorf("failed to execute the prepopulate script: %w", err)
	}
	return nil
}

func teardown(r *setupResult) {
	if err := r.Pool.Purge(r.PostgresContainer); err != nil {
		log.Printf("failed to purge the Postgres container: %v", err)
	}
}

func TestGetVideosByCaption(t *testing.T) {
	conn, err := getDBConnector()
	if err != nil {
		t.Fatalf("failed to get a connector to the DB: %v", err)
	}

	captionTestSubstring := "test_get_videos_by_caption"
	videosForTest := []storage.FoundVideo{
		{
			Caption:  "5 stars test_get_videos_by_caption very interesting",
			URI:      "https://someURI.org/some-test-video",
			Location: "\\asdf.qwerty",
		},
		{
			Caption:  "test_get_videos_by_caption very interesting, 4 stars",
			URI:      "https://one-more-uri.org/one-more-test",
			Location: "\\some-location.adsf",
		},
		{
			Caption:  "6 stars, must see, test_get_videos_by_caption",
			URI:      "https://two-more-uri.org/one-more-test",
			Location: "\\some-new-location.adsf",
		},
	}

	// todo скорректировать запрос для пополнения БД, наполненную тестовыми данными

	batch := &pgx.Batch{}
	const query = `INSERT INTO videos (user_id, location, uri, res, caption, description, created_at, updated_at)
		VALUES(
			20, 
			$1, $2 ,'240p', $3, 
			'Tempore laborum deleniti et officia ab et omnis. Possimus perferendis maxime itaque in. Vel hic suscipit temporibus et accusamus odit.',
			'1973-04-01 23:06:00',
			'2007-07-03 19:57:54'
		)`
	for _, v := range videosForTest {
		batch.Queue(
			query,
			v.Location,
			v.URI,
			v.Caption,
		)
	}
	if _, err := conn.SendBatch(context.Background(), batch).Exec(); err != nil {
		t.Fatalf("failed to create DB data: %v", err)
	}

	db, err := storage.NewDB(getConnectionString())
	if err != nil {
		t.Fatalf("failed to create a DB object: %v", err)
	}
	videos, err := db.GetVideosByCaption(context.Background(), captionTestSubstring)
	if err != nil {
		t.Fatalf("GetVideosByCaption failed: %v", err)
	}
	if len(videosForTest) != len(videos) {
		t.Fatalf("wrong number of found videos: expected %d, got %d", len(videosForTest), len(videos))
	}

	// todo попытаться понять, для чего вообще эти строки
	sort.Slice(videosForTest, func(i int, j int) bool {
		return videosForTest[i].Caption < videosForTest[j].Caption
	})
	sort.Slice(videos, func(i int, j int) bool {
		return videos[i].Caption < videos[j].Caption
	})
	for i, tst := range videosForTest {
		v := *videos[i]
		if tst != v {
			t.Fatalf("expected object %v is not equal to the actual object %v", tst, v)
		}
	}
}

func getDBConnector() (*pgxpool.Pool, error) {
	log.Println(composeConnectionString())
	cfg, err := pgxpool.ParseConfig(composeConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to create the PGX pool config from connection string: %w", err)
	}
	cfg.ConnConfig.ConnectTimeout = time.Second * 1
	db, err := pgxpool.ConnectConfig(context.Background(), cfg)
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
	return cfg, nil
}

func getConnectionString() *storage.ConnString {
	return &storage.ConnString{
		Host:     DB_HOST,
		Port:     DB_PORT,
		User:     DB_USER,
		Password: DB_PASSWORD,
		DBName:   DB_NAME,
	}
}

func composeConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", DB_USER, url.QueryEscape(DB_PASSWORD), DB_HOST, DB_PORT, DB_NAME)
}
