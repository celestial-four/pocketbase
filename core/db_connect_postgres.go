package core

import (
	"fmt"

	"github.com/pocketbase/dbx"
)

// PostgreSQLDBConnect creates a PostgreSQL database connection.
//
// **WARNING: PostgreSQL support is currently experimental and not production-ready.**
// This function is provided as a foundation for future PostgreSQL support.
// It has not been thoroughly tested and may not work correctly with PocketBase's
// data model, migrations, and SQLite-specific queries.
//
// Use at your own risk. For production use, stick with SQLite.
func PostgreSQLDBConnect(dbURL string) (*dbx.DB, error) {
	return nil, fmt.Errorf("PostgreSQL support is not yet implemented - SQLite is the only supported database")
}
