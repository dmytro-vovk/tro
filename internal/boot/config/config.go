package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type WebServer struct {
	Listen string `json:"listen"`
	TLS    TLS    `json:"tls"`
}

type TLS struct {
	Enabled  bool   `json:"enabled"`
	HostName string `json:"host_name"`
	CertDir  string `json:"cert_dir"`
}

type API struct {
	AuthMethod string `json:"auth_method"`
	Listen     string `json:"listen"`
}

type Database struct {
	DriverName string `json:"driver_name"`
}

func DefaultViper() *viper.Viper {
	v := viper.New()

	for key, value := range map[string]interface{}{
		"logger.level":                    logrus.InfoLevel.String(),
		"logger.timestamp_format":         "2006-01-02 15:04:05",
		"logger.formatter.line":           false,
		"logger.formatter.package":        false,
		"logger.formatter.file":           false,
		"logger.formatter.base_name_only": false,
		"logger.rotor.filename":           "logs/app.log",
		"logger.rotor.max_size":           100,
		"logger.rotor.max_age":            0,
		"logger.rotor.max_backups":        0,
		"logger.rotor.compress":           false,
	} {
		v.SetDefault(key, value)
	}

	return v
}
