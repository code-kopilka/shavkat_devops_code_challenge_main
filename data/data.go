package data

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/LogicGateTech/devops-code-challenge/conf"
	_ "github.com/mattn/go-sqlite3"
)

const (
	SqliteCmd = "sqlite3"
)

type DB struct {
	conf *conf.Conf
	log  *slog.Logger

	Conn *sql.DB
}

func (db *DB) Dsn() string {
	// Use configurable database path
	dbPath := db.conf.DatabasePath
	if dbPath == "" {
		dbPath = "data.db"
	}
	// Only add auth if password is set
	if db.conf.Password != "" {
		return fmt.Sprintf("file:%s?cache=shared&_auth&_auth_user=admin&_auth_pass=%s&_auth_crypt=sqlite_crypt", dbPath, db.conf.Password)
	}
	return fmt.Sprintf("file:%s?cache=shared", dbPath)
}

func (db *DB) Bootstrap() error {
	if db.Conn == nil {
		return fmt.Errorf("missing db conn. call Open() first?")
	}
	if _, err := db.Conn.Exec(CreateUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func Open() (*DB, error) {
	var err error
	db := &DB{}

	if db.conf, err = conf.New(); err != nil {
		return nil, err
	}

	// Create logger based on configuration
	db.log = conf.NewLogger(db.conf)
	// Don't log DSN as it contains sensitive password information
	db.log.Info("data: opening database connection", "path", db.conf.DatabasePath)
	if db.Conn, err = sql.Open(SqliteCmd, db.Dsn()); err != nil {
		return nil, err
	}
	// Configure connection pool
	db.Conn.SetMaxOpenConns(25)
	db.Conn.SetMaxIdleConns(5)
	db.Conn.SetConnMaxLifetime(5 * time.Minute) // Connections expire after 5 minutes
	db.Conn.SetConnMaxIdleTime(2 * time.Minute) // Idle connections expire after 2 minutes
	return db, nil
}
