package database

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mamcer/gostalgia/internal/config"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewMySQLDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.NOSTALGIA_CONNECTION_STRING), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func RunMigrations(cfg *config.Config) {
	db, err := sql.Open("mysql", cfg.NOSTALGIA_CONNECTION_STRING)
	if err != nil {
		log.Fatalf("Failed to open sql connection for migrations: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
