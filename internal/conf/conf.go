package conf

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ExternalURL string `yaml:"external_url"`
	Web         struct {
		Port int `yaml:"port"`
	} `yaml:"web"`
	Proxy Proxy `yaml:"proxy"`
	SSHD  struct {
		Port int `yaml:"port"`
	} `yaml:"sshd"`
	Database         *Database         `yaml:"database"`
	IdentityProvider *IdentityProvider `yaml:"identity_provider"`
}

type Proxy struct {
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
	Domain string `yaml:"domain"`
	TCP    struct {
		Domain    string `yaml:"domain"`
		PortStart int    `yaml:"port_start"`
		PortEnd   int    `yaml:"port_end"`
	} `yaml:"tcp"`
}

// Supported database backends.
const (
	DatabaseTypePostgres = "postgres"
	DatabaseTypeMySQL    = "mysql"
	DatabaseTypeSQLite   = "sqlite"
)

type Database struct {
	// Type is the database backend: "postgres", "mysql", or "sqlite".
	// Defaults to "postgres" when empty.
	Type string `yaml:"type"`

	// Path is the file path for sqlite. Ignored for other backends.
	Path string `yaml:"path"`

	// Host/Port/User/Password/Database are used by postgres and mysql.
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type IdentityProvider struct {
	Type         string `yaml:"type"`
	DisplayName  string `yaml:"display_name"`
	Issuer       string `yaml:"issuer"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	FieldMapping struct {
		Identifier  string `yaml:"identifier"`
		DisplayName string `yaml:"display_name"`
		Email       string `yaml:"email"`
	} `yaml:"field_mapping"`
	RequiredDomain string `yaml:"required_domain"`
}

// Load returns the config loaded from the given path.
func Load(configPath string) (*Config, error) {
	p, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}

	var config Config

	if config.Proxy.TCP.PortStart <= 0 {
		config.Proxy.TCP.PortStart = 10000
	}

	if config.Proxy.TCP.PortEnd <= 0 {
		config.Proxy.TCP.PortEnd = 15000
	}

	err = yaml.Unmarshal(p, &config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	config.ExternalURL = strings.TrimSuffix(config.ExternalURL, "/")

	if db := config.Database; db != nil {
		if db.Type == "" {
			db.Type = DatabaseTypePostgres
		}
		switch db.Type {
		case DatabaseTypeSQLite:
			if db.Path == "" {
				db.Path = "pgrokd.db"
			}
		case DatabaseTypePostgres, DatabaseTypeMySQL:
			if db.Host == "" || db.User == "" || db.Database == "" {
				return nil, errors.Errorf("database type %q requires host, user, and database", db.Type)
			}
		default:
			return nil, errors.Errorf("unsupported database type %q (expected one of: postgres, mysql, sqlite)", db.Type)
		}
	}

	if idp := config.IdentityProvider; idp != nil {
		if idp.RequiredDomain != "" && idp.FieldMapping.Email == "" {
			return nil, errors.New("cannot require email domain without field mapping for email")
		}
	}
	return &config, nil
}
