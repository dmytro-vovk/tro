package repository

import (
	"fmt"
	"github.com/spf13/viper"

	"github.com/dmytro-vovk/tro/internal/api/model"
	"github.com/dmytro-vovk/tro/internal/api/repository/mysql"
	"github.com/dmytro-vovk/tro/internal/api/repository/postgres"
)

type Storage interface {
	Close() error
	DriverName() string
}

type Authorization interface {
	CreateUser(user model.User) (int, error)
	GetUser(username, password string) (model.User, error)
}

type Repository interface {
	Storage
	Authorization
}

func New(v *viper.Viper) (Repository, error) {
	errFormat := "can't unmarshall %q config: %w"
	switch driver := v.GetString("database.driver_name"); driver {
	case "mysql":
		var config mysql.Config
		if err := v.Unmarshal(&config); err != nil {
			return nil, fmt.Errorf(errFormat, driver, err)
		}

		return mysql.New(config)
	case "postgres":
		var config postgres.Config
		if err := v.Unmarshal(&config); err != nil {
			return nil, fmt.Errorf(errFormat, driver, err)
		}

		return postgres.New(config)
	default:
		return nil, fmt.Errorf("unknown database driver name %q", driver)
	}
}
