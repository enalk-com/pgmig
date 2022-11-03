package pgmig

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var pgmShellPrefix = ">>> "

type Migrate struct {
	c *Config
}

func New(cp string) (*Migrate, error) {
	c, err := NewConfig(cp)
	if err != nil {
		return nil, err
	}
	return &Migrate{
		c: c,
	}, nil
}

func shellOut(value []byte) {
	s := fmt.Sprintf("%s%s", pgmShellPrefix, string(value))
	fmt.Println(s)
}

func (m Migrate) newMigrationFile() error {
	shellOut([]byte("Please enter a migration title"))
	title, err := readUserInput()
	if err != nil {
		fmt.Printf("Errored\n")
		return err
	}

	ts := time.Now().Unix()
	uptitle := fmt.Sprintf("%d_%s.up.sql", ts, title)
	downtitle := fmt.Sprintf("%d_%s.down.sql", ts, title)

	fmt.Printf(`

Creating:
UP: %s
DOWN: %s

`, uptitle, downtitle)
	pup := fmt.Sprintf("%s/%s", strings.TrimRight(m.c.MigrationsDir, "/"), uptitle)
	pdown := fmt.Sprintf("%s/%s", strings.TrimRight(m.c.MigrationsDir, "/"), downtitle)

	_, err = os.Create(pup)
	if err != nil {
		fmt.Printf("Errored\n")
		return err
	}

	_, err = os.Create(pdown)
	if err != nil {
		fmt.Printf("Errored\n")
		return err
	}

	fmt.Printf("Created\n\n")
	return nil
}

func (m Migrate) forceVersion() error {
	shellOut([]byte("Enter the version to force:"))
	v, err := readUserInput()
	if err != nil {
		fmt.Printf("Errored\n")
		return nil
	}

	version, err := strconv.Atoi(v)
	if err != nil {
		fmt.Printf("Errored\n")
		return err
	}

	err = m.c.instance.Force(version)
	if err != nil {
		fmt.Printf("Errored\n")
		return err
	}
	fmt.Printf("Forced Version: %d\n\n", version)
	return nil
}

func (m Migrate) upN() error {
	shellOut([]byte("Enter number of steps to migrate up. Should be a +ve interger greater than 0:"))
	v, err := readUserInput()
	if err != nil {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	if err := m.c.UpN(n); err != nil {
		return err
	}
	return nil
}

func (m Migrate) downN() error {
	shellOut([]byte("Enter number of steps to migrate down. Should be a +ve interger greater than 0:"))
	v, err := readUserInput()
	if err != nil {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	if err := m.c.DownN(n); err != nil {
		return err
	}
	return nil
}

func readUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdout)
	r, err := reader.ReadString('\n')
	if err == nil {
		r = strings.Trim(r, " \t\n")
	}
	return r, err
}

func (m Migrate) Wizard() {
	instructions := `Welcome to the postgres migrations wizard.
Choose an option below:
1. New Migration File
2. Migrate Up
3. Migrate Up (n) Steps
4. Migrate Down
5. Migrate Down (n) Steps
6. ReMigrate. Down -> Up
7. Force Version
=============================
0. Quit`

	shellOut([]byte(instructions))
	option, err := readUserInput()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if option == "0" {
		return
	}

	var serr error
	switch option {
	case "1":
		serr = m.newMigrationFile()
	case "2":
		serr = m.c.Up()
	case "3":
		serr = m.upN()
	case "4":
		serr = m.c.Down()
	case "5":
		serr = m.downN()
	case "6":
		serr = m.c.DownN(1)
		if serr == nil {
			serr = m.c.UpN(1)
		}
	case "7":
		serr = m.forceVersion()
	default:
		fmt.Printf("\nInvalid option \"%s\"\n", option)
	}
	if serr != nil {
		fmt.Println(serr.Error())
	}
}
