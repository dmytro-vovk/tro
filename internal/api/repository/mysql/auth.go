package mysql

import (
	"fmt"
	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/sirupsen/logrus"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Username string `mapstructure:"MYSQL_USERNAME"`
	Password string `mapstructure:"MYSQL_PASSWORD"`
	Host     string `mapstructure:"MYSQL_HOST"`
	Port     string `mapstructure:"MYSQL_PORT"`
	Database string `mapstructure:"MYSQL_DATABASE"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}

type storage struct {
	db *sqlx.DB
}

func New(c Config) (*storage, error) {
	db, err := sqlx.Open("mysql", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("error conecting to database: %w", err)
	}

	for {
		if err := db.Ping(); err != nil {
			logrus.Printf("Error pinging database: [%T] %s", err, err)
			time.Sleep(time.Second)

			continue
		}

		break
	}

	return &storage{db: db}, nil
}

func (s *storage) Close() error { return s.db.Close() }

func (s *storage) DriverName() string { return s.db.DriverName() }

func (s *storage) CreateUser(user model.User) (int, error) {
	return 0, nil
}

func (s *storage) GetUser(username, password string) (model.User, error) {
	return model.User{}, nil
}
