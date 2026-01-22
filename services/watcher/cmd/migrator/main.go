package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func main() {
	var (
		dsn             string
		migrationsPath  string
		migrationsTable string
		down            bool
	)

	flag.StringVar(&dsn, "dsn", "", "postgres dsn")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.BoolVar(&down, "down", false, "rollback migrations")
	flag.Parse()

	if dsn == "" {
		log.Fatal("dsn is required")
	}
	if migrationsPath == "" {
		log.Fatal("migrations-path is required")
	}

	sourceURL := "file://" + migrationsPath
	databaseURL := fmt.Sprintf(
		"%s&x-migrations-table=%s",
		dsn,
		migrationsTable,
	)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	if down {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("no migrations to rollback")
				return
			}
			log.Fatal(err)
		}
		log.Println("migrations rolled back")
		return
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")
			return
		}
		log.Fatal(err)
	}

	log.Println("migrations applied successfully")
}
