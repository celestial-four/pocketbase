package core

import (
	"github.com/pocketbase/dbx"

	// Register pgx PostgreSQL driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgreSQLDBConnect creates a PostgreSQL database connection.
//
// The dbURL should be a valid PostgreSQL connection string, e.g.:
//   - "postgres://user:password@localhost:5432/dbname"
//   - "postgres://user:password@localhost:5432/dbname?sslmode=disable"
//   - "host=localhost port=5432 user=user password=password dbname=dbname sslmode=disable"
//
// **WARNING: PostgreSQL support is experimental.**
// While basic functionality is implemented, some advanced features may not work
// as expected. Thoroughly test in a non-production environment first.
//
// For production use with simpler deployments, SQLite is still recommended.
//
// Note: Connection pool settings (MaxOpenConns, MaxIdleConns, ConnMaxLifetime)
// are configured via BaseAppConfig and applied after connection is established.
// The dbURL may contain credentials - ensure proper handling when logging errors.
func PostgreSQLDBConnect(dbURL string) (*dbx.DB, error) {
	db, err := dbx.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
