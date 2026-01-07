package core

import (
	"fmt"
	"os"
	"strings"
)

// DatabaseType represents the type of database backend.
type DatabaseType string

const (
	// DatabaseTypeSQLite represents SQLite database (default).
	DatabaseTypeSQLite DatabaseType = "sqlite"
	// DatabaseTypePostgreSQL represents PostgreSQL database.
	// Note: PostgreSQL support is currently experimental and not fully implemented.
	DatabaseTypePostgreSQL DatabaseType = "postgres"
)

// DatabaseConfig holds database configuration.
// This struct provides a foundation for supporting multiple database backends.
// Currently, only SQLite is fully supported.
type DatabaseConfig struct {
	// Type is the database type (sqlite or postgres).
	Type DatabaseType

	// URL is the database connection URL.
	// For PostgreSQL: connection string (e.g., "postgres://user:pass@host:5432/dbname")
	// For SQLite: this field is ignored (uses DataDir-based paths)
	URL string
}

// GetDatabaseConfigFromEnv reads database configuration from environment variables.
//
// Environment variables:
//   - PB_DB_TYPE: database type ("sqlite" or "postgres"). Defaults to "sqlite".
//   - PB_DB_URL: database connection URL (required for postgres, ignored for sqlite).
//
// This function always returns a valid DatabaseConfig with SQLite as the default type.
func GetDatabaseConfigFromEnv() *DatabaseConfig {
	dbType := os.Getenv("PB_DB_TYPE")
	dbURL := os.Getenv("PB_DB_URL")

	// Default to SQLite if not specified
	if dbType == "" {
		dbType = string(DatabaseTypeSQLite)
	}

	return &DatabaseConfig{
		Type: DatabaseType(strings.ToLower(dbType)),
		URL:  dbURL,
	}
}

// Validate checks if the database configuration is valid.
// Returns an error if the configuration is invalid or unsupported.
func (dc *DatabaseConfig) Validate() error {
	switch dc.Type {
	case DatabaseTypeSQLite:
		// SQLite doesn't require URL (will use DataDir-based paths)
		return nil
	case DatabaseTypePostgreSQL:
		// PostgreSQL requires a connection URL
		if dc.URL == "" {
			return fmt.Errorf("PostgreSQL requires a connection URL (set PB_DB_URL environment variable or --dbURL flag)")
		}
		return nil
	default:
		return fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres)", dc.Type)
	}
}

// IsPostgreSQL returns true if the database type is PostgreSQL.
func (dc *DatabaseConfig) IsPostgreSQL() bool {
	return dc.Type == DatabaseTypePostgreSQL
}

// IsSQLite returns true if the database type is SQLite (the default).
func (dc *DatabaseConfig) IsSQLite() bool {
	return dc.Type == DatabaseTypeSQLite || dc.Type == ""
}
