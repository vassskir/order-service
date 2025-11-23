package main

import (
	"L0/internal/config"
	"L0/internal/logger"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	logger.Init("migrate")
	cfg := config.Load()

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|version]")
		os.Exit(1)
	}

	m, err := migrate.New("file://migrations", cfg.Database.GetDBConnectionString())
	if err != nil {
		logger.Error("Migration failed", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Error("Up failed", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			logger.Error("Down failed", err)
			os.Exit(1)
		}
		fmt.Println("Migrations rolled back")
	case "version":
		v, d, err := m.Version()
		if err != nil {
			fmt.Println("No migrations applied")
		} else {
			fmt.Printf("Version: %d, Dirty: %t\n", v, d)
		}
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}
