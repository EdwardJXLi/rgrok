package database

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/flamego/flamego"
	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/EdwardJXLi/rgrok/internal/conf"
)

// DB is the database handle.
type DB struct {
	*gorm.DB
}

// New returns a new database handle with given configuration.
func New(logWriter io.Writer, config *conf.Database) (*DB, error) {
	if config == nil {
		return nil, errors.New("no database config provided")
	}

	level := logger.Info
	if flamego.Env() == flamego.EnvTypeProd {
		level = logger.Warn
	}

	// NOTE: AutoMigrate does not respect logger passed in gorm.Config.
	logger.Default = logger.New(
		&gormLogger{
			Logger: log.NewWithOptions(
				logWriter,
				log.Options{
					TimeFormat:      time.DateTime,
					Level:           log.DebugLevel,
					Prefix:          "gorm",
					ReportTimestamp: true,
				},
			),
		},
		logger.Config{
			SlowThreshold: 1000 * time.Millisecond,
			LogLevel:      level,
		},
	)

	dialector, err := dialectorFor(config)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(
		dialector,
		&gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return time.Now().UTC().Truncate(time.Microsecond)
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get underlying *sql.DB")
	}
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(30)
	sqlDB.SetConnMaxLifetime(time.Minute)

	err = db.AutoMigrate(&Principal{}, &HostKey{})
	if err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}
	return &DB{db}, nil
}

func dialectorFor(config *conf.Database) (gorm.Dialector, error) {
	switch config.Type {
	case conf.DatabaseTypePostgres:
		return postgres.Open(fmt.Sprintf(
			"user='%s' password='%s' host='%s' port='%d' dbname='%s' search_path='public' application_name='rgrokd' client_encoding=UTF8",
			config.User, config.Password, config.Host, config.Port, config.Database,
		)), nil
	case conf.DatabaseTypeMySQL:
		return mysql.Open(fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC&charset=utf8mb4",
			config.User, config.Password, config.Host, config.Port, config.Database,
		)), nil
	case conf.DatabaseTypeSQLite:
		return sqlite.Open(config.Path), nil
	}
	return nil, errors.Errorf("unsupported database type %q", config.Type)
}

// gormLogger is a wrapper of io.Writer for the GORM's logger.Writer.
type gormLogger struct {
	*log.Logger
}

func (l *gormLogger) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	print := l.Debug
	if strings.Contains(msg, "[error]") {
		print = l.Error
	}
	print(msg)
}
