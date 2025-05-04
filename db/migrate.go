package db

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

func RunMigrations() {
	driver, err := postgres.WithInstance(DB.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Migration driver init failed: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Migration instance creation failed")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration Failed: %v", err)
	}

	log.Println("Migrations applied successfully")
}
