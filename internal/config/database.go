package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewDatabase(env *Env) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		env.DBUsername,
		env.DBPassword,
		env.DBHost,
		env.DBPort,
		env.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql connection: %w", err)
	}

	db.SetConnMaxLifetime(time.Second * time.Duration(env.DBConnMaxLifetime))
	db.SetMaxOpenConns(env.DBMaxOpenConn)
	db.SetMaxIdleConns(env.DBMaxIdleConn)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
