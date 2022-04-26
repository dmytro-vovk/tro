package config

import (
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	Level            string            `mapstructure:"level"`
	TimestampFormat  string            `mapstructure:"timestamp_format"`
	RuntimeFormatter *runtimeFormatter `mapstructure:"formatter"`
	Rotor            *rotor            `mapstructure:"rotor"`
}

type runtimeFormatter struct {
	Line         bool `mapstructure:"line"`
	Package      bool `mapstructure:"package"`
	File         bool `mapstructure:"file"`
	BaseNameOnly bool `mapstructure:"base_name_only"`
}

type rotor struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	LocalTime  bool   `mapstructure:"local_time"`
	Compress   bool   `mapstructure:"compress"`
}

func (l *Logger) Formatter(child logrus.Formatter) logrus.Formatter {
	f := l.RuntimeFormatter
	if f == nil || !(f.Line || f.Package || f.File) {
		return child
	}

	return &runtime.Formatter{
		ChildFormatter: child,
		Line:           f.Line,
		Package:        f.Package,
		File:           f.File,
		BaseNameOnly:   f.BaseNameOnly,
	}
}

func (l *Logger) FieldMap() logrus.FieldMap {
	return logrus.FieldMap{
		logrus.FieldKeyMsg: "message",
	}
}
