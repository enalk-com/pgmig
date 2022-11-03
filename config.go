package pgmig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	migrate "github.com/golang-migrate/migrate/v4"
)

func validConfigurationFile(p string) bool {
	return strings.ToLower(filepath.Ext(p)) != ".json"
}

type Config struct {
	MigrationsDir string `toml:"migrations_dir"`
	Host          string `toml:"host"`
	Port          string `toml:"port"`
	User          string `toml:"user"`
	Pwd           string `toml:"pwd"`
	DB            string `toml:"db"`
	SSL_MODE      string `toml:"sslmode"`
	SSL_CERT      string `toml:"sslcert"`
	SSL_KEY       string `toml:"sslkey"`
	instance      *migrate.Migrate
}

func (c Config) validate() error {
	valid := func(s string) bool {
		return !strings.ContainsRune(s, '\\') && !strings.ContainsRune(s, '\'')
	}

	es := "single quotes and baclslashes are not allowed"
	if !valid(c.Host) {
		return fmt.Errorf("invalid host. %s", es)
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return err
	}

	if !valid(c.User) {
		return fmt.Errorf("invalid user. %s", es)
	}

	if !valid(c.DB) {
		return fmt.Errorf("invalid db. %s", es)
	}

	if c.SSL_MODE != "disable" && c.SSL_MODE != "allow" && c.SSL_MODE != "prefer" &&
		c.SSL_MODE != "require" && c.SSL_MODE != "verify-ca" && c.SSL_MODE != "verify_full" {
		return fmt.Errorf("invalid value \"%s\" for sslmode", c.SSL_MODE)
	}

	if !valid(c.SSL_CERT) {
		return fmt.Errorf("invalid sslcert. %s", es)
	}

	if !valid(c.SSL_KEY) {
		return fmt.Errorf("invalid sslkey. %s", es)
	}
	return nil
}

func readFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c Config) ensureMigrationDirectory() error {
	if r, err := os.Stat(c.MigrationsDir); err == nil {
		if r.IsDir() {
			return nil
		}

		err = os.Remove(c.MigrationsDir)
		if err != nil {
			return err
		}

		err = os.MkdirAll(c.MigrationsDir, os.ModeDir)
		if err != nil {
			return err
		}
		return nil
	} else if os.IsNotExist(err) {

		err := os.MkdirAll(c.MigrationsDir, os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("could not ensure migrations directory at: %s", c.MigrationsDir)
}

func NewConfig(p string) (*Config, error) {
	if !validConfigurationFile(p) {
		return nil, fmt.Errorf("configuration file path invalid")
	}
	bs, err := readFile(p)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := toml.Unmarshal(bs, &c); err != nil {
		return nil, err
	}
	if c.SSL_MODE == "" {
		c.SSL_MODE = "disable"
	}

	// Since DSN mode
	c.Pwd = strings.ReplaceAll(c.Pwd, "\\", "\\\\")
	c.Pwd = strings.ReplaceAll(c.Pwd, "'", "\\'")

	if err := c.validate(); err != nil {
		return nil, err
	}
	if err := c.ensureMigrationDirectory(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c Config) ConnURL() string {
	s := []string{}
	if c.User != "" {
		s = append(s, "user="+c.User)
	}
	if c.Pwd != "" {
		s = append(s, "password="+c.Pwd)
	}
	s = append(s, "host="+c.Host, "port="+c.Port, "dbname="+c.DB, "sslmode="+c.SSL_MODE)
	if c.SSL_CERT != "" {
		s = append(s, "sslcert="+c.SSL_CERT)
	}
	if c.SSL_KEY != "" {
		s = append(s, "sslkey="+c.SSL_KEY)
	}
	return strings.Join(s, " ")
}

func (c Config) UpN(n int) error {
	if n <= 0 {
		return fmt.Errorf("argument 'n' must be a +ve integer")
	}
	return c.instance.Steps(n)
}

func (c Config) DownN(n int) error {
	if n <= 0 {
		return fmt.Errorf("argument 'n' must be a +ve integer")
	}
	return c.instance.Steps(n * -1)
}

func (c Config) Up() error {
	err := c.instance.Up()
	if err != nil {
		return err
	}
	return nil
}

func (c Config) Down() error {
	err := c.instance.Down()
	if err != nil {
		return err
	}
	return nil
}
