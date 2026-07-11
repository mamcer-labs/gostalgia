package repository

import (
	"context"
	"testing"

	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SetupMySQLContainer creates a real MySQL container for integration tests.
// Note: This requires Docker to be running in the environment.
func SetupMySQLContainer(ctx context.Context, t *testing.T) (*gorm.DB, func()) {
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("nostalgia_test"),
		mysql.WithUsername("test"),
		mysql.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	// GORM connection
	db, err := gorm.Open(gormmysql.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %s", err)
	}

	// Migrate
	if err := db.AutoMigrate(&domain.NTag{}, &domain.NFile{}, &domain.NDirectory{}, &domain.NScan{}, &domain.NFileNode{}); err != nil {
		t.Fatalf("failed to migrate database: %s", err)
	}

	cleanUp := func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	return db, cleanUp
}

func TestWithRealMySQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
