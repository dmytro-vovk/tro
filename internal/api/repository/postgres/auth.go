package postgres

import (
	"fmt"

	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/jmoiron/sqlx"
)

const usersTable = "users"

type Config struct {
	Username string `mapstructure:"POSTGRESQL_USERNAME"`
	Password string `mapstructure:"POSTGRESQL_PASSWORD"`
	Host     string `mapstructure:"POSTGRESQL_HOST"`
	Port     string `mapstructure:"POSTGRESQL_PORT"`
	Database string `mapstructure:"POSTGRESQL_DATABASE"`
	SSLMode  string `mapstructure:"POSTGRESQL_SSLMODE"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

type storage struct {
	db *sqlx.DB
}

func New(c Config) (*storage, error) {
	db, err := sqlx.Open("postgres", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("error conecting to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &storage{db: db}, nil
}

func (s *storage) Close() error { return s.db.Close() }

func (s *storage) DriverName() string { return s.db.DriverName() }

func (s *storage) CreateUser(user model.User) (int, error) {
	var id int
	query := fmt.Sprintf("INSERT INTO %s (name, username, password_hash) values ($1, $2, $3) RETURNING id", usersTable)
	row := s.db.QueryRow(query, user.Name, user.Username, user.Password)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *storage) GetUser(username, password string) (model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT id FROM %s WHERE username=$1 AND password_hash=$2", usersTable)
	err := s.db.Get(&user, query, username, password)

	return user, err
}
