package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	postgresDriverName = "postgres"
)

type Storage struct {
	Db *sqlx.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sqlx.Connect(postgresDriverName, dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{
		Db: db,
	}, nil
}

func (s *Storage) RunMigrations(migrationsPath string) error {
	if err := goose.SetDialect(postgresDriverName); err != nil {
		return err
	}

	if err := goose.Up(s.Db.DB, migrationsPath); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Close() error {
	return s.Db.Close()
}
