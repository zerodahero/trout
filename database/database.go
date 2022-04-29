package database

import (
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var fs embed.FS

var db *gorm.DB

func InitDB(dbPath string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return errors.Wrap(err, "failed to open sqlite DB")
	}

	runMigrations(db)

	return nil
}

func runMigrations(db *gorm.DB) error {
	sqlDb, err := db.DB()
	if err != nil {
		return fmt.Errorf("unable to get root db instance %s", err)
	}
	driver, err := sqlite3.WithInstance(sqlDb, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("creating sqlite3 db driver failed %s", err)
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations %s", err)
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("initializing db migration failed %s", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrating database failed %s", err)
	}

	return nil
}
