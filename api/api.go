package api

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/LogicGateTech/devops-code-challenge/conf"
	"github.com/LogicGateTech/devops-code-challenge/data"
)

type API struct {
	log  *slog.Logger
	conf *conf.Conf
	db   *data.DB
}

// Close closes all resources held by the API
func (api *API) Close() error {
	if api.db != nil && api.db.Conn != nil {
		return api.db.Conn.Close()
	}
	return nil
}

// HealthCheck checks if the database is accessible
func (api *API) HealthCheck() error {
	if api.db == nil || api.db.Conn == nil {
		return fmt.Errorf("database not initialized")
	}
	return api.db.Conn.Ping()
}

func New() (*API, error) {
	var err error
	api := &API{}

	if api.conf, err = conf.New(); err != nil {
		return nil, err
	}

	// Create logger based on configuration
	api.log = conf.NewLogger(api.conf)
	if api.db, err = data.Open(); err != nil {
		return nil, err
	}
	// Verify database connection
	if err = api.db.Conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if err = api.db.Bootstrap(); err != nil {
		return nil, fmt.Errorf("failed to bootstrap database: %w", err)
	}
	return api, nil
}

func (api *API) Signup(username, password string) error {
	// Hash password before storing
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	created := time.Now().Unix()
	result, err := api.db.Conn.Exec(data.CreateUser, username, hashedPassword, created)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows != int64(1) {
		return fmt.Errorf("failed to create user: %s", username)
	}
	return nil
}

func (api *API) Reset(username, password string) error {
	// Hash password before storing
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	updated := time.Now().Unix()
	result, err := api.db.Conn.Exec(data.ResetPassword, hashedPassword, updated, username)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows != int64(1) {
		return fmt.Errorf("user not found: %s", username)
	}
	return nil
}
