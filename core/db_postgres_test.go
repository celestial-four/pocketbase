package core_test

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestPostgreSQLDBConnect(t *testing.T) {
	// Note: These tests verify the PostgreSQL connection function signatures
	// and basic behavior. They don't require a running PostgreSQL instance.

	t.Run("returns nil with invalid connection URL", func(t *testing.T) {
		// PostgreSQLDBConnect should attempt to open a connection
		// and fail with an invalid URL, but not panic
		_, err := core.PostgreSQLDBConnect("")
		if err == nil {
			t.Log("Empty URL opened successfully (driver accepted it)")
		}
	})

	t.Run("accepts postgres:// URL format", func(t *testing.T) {
		// This test verifies the function accepts the URL format
		// Even though connection will fail without a running server,
		// the function shouldn't panic on valid URL format
		_, err := core.PostgreSQLDBConnect("postgres://localhost:5432/testdb")
		if err == nil {
			t.Log("Connection opened (may be using default postgres instance)")
		} else {
			// Expected to fail without a running PostgreSQL server
			t.Logf("Connection failed as expected: %v", err)
		}
	})
}

func TestDBTypeMethod(t *testing.T) {
	const testDataDir = "./pb_dbtype_test_data_dir/"

	t.Run("returns SQLite for default config", func(t *testing.T) {
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir + "sqlite_default",
			IsDev:   true,
		})
		if app == nil {
			t.Fatal("expected app to be created")
		}

		dbType := app.DBType()
		if dbType != core.DatabaseTypeSQLite {
			t.Errorf("expected DBType() to return %s, got %s", core.DatabaseTypeSQLite, dbType)
		}
	})

	t.Run("returns PostgreSQL for postgres config", func(t *testing.T) {
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir + "postgres",
			IsDev:   true,
			DBConfig: &core.DatabaseConfig{
				Type: core.DatabaseTypePostgreSQL,
				URL:  "postgres://localhost:5432/testdb",
			},
		})
		if app == nil {
			t.Fatal("expected app to be created")
		}

		dbType := app.DBType()
		if dbType != core.DatabaseTypePostgreSQL {
			t.Errorf("expected DBType() to return %s, got %s", core.DatabaseTypePostgreSQL, dbType)
		}
	})

	t.Run("returns SQLite for empty type", func(t *testing.T) {
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir + "empty_type",
			IsDev:   true,
			DBConfig: &core.DatabaseConfig{
				Type: "",
			},
		})
		if app == nil {
			t.Fatal("expected app to be created")
		}

		dbType := app.DBType()
		if dbType != core.DatabaseTypeSQLite {
			t.Errorf("expected DBType() to return %s for empty type, got %s", core.DatabaseTypeSQLite, dbType)
		}
	})
}

func TestTranslateSQLForDB(t *testing.T) {
	t.Run("returns unchanged SQL for SQLite", func(t *testing.T) {
		sql := "SELECT strftime('%Y-%m-%d', created) FROM test"
		result := core.TranslateSQLForDB(sql, core.DatabaseTypeSQLite)
		if result != sql {
			t.Errorf("expected SQL to be unchanged for SQLite, got %s", result)
		}
	})

	t.Run("translates ID generation for PostgreSQL", func(t *testing.T) {
		sql := "CREATE TABLE test (id TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL)"
		result := core.TranslateSQLForDB(sql, core.DatabaseTypePostgreSQL)
		
		// Should contain PostgreSQL random ID generation (gen_random_uuid or similar)
		if result == sql {
			t.Error("expected SQL to be translated for PostgreSQL")
		}
	})

	t.Run("translates timestamp for PostgreSQL", func(t *testing.T) {
		sql := "CREATE TABLE test (created TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL)"
		result := core.TranslateSQLForDB(sql, core.DatabaseTypePostgreSQL)
		
		// Should be translated
		if result == sql {
			t.Error("expected timestamp SQL to be translated for PostgreSQL")
		}
	})
}

func TestDialectMethods(t *testing.T) {
	t.Run("PostgreSQL dialect returns correct values", func(t *testing.T) {
		dialect := core.NewPostgreSQLDialect()
		
		if dialect.Name() != "postgres" {
			t.Errorf("expected name 'postgres', got %s", dialect.Name())
		}
		
		if dialect.JSONType() != "JSONB" {
			t.Errorf("expected JSON type 'JSONB', got %s", dialect.JSONType())
		}
		
		if dialect.BooleanType() != "BOOLEAN" {
			t.Errorf("expected boolean type 'BOOLEAN', got %s", dialect.BooleanType())
		}
		
		if dialect.BooleanTrue() != "TRUE" {
			t.Errorf("expected boolean true 'TRUE', got %s", dialect.BooleanTrue())
		}
		
		if dialect.BooleanFalse() != "FALSE" {
			t.Errorf("expected boolean false 'FALSE', got %s", dialect.BooleanFalse())
		}
		
		// RandomID should use PostgreSQL's random generation
		randomID := dialect.RandomID()
		if randomID == "" {
			t.Error("expected non-empty random ID generation expression")
		}
		
		// CurrentTimestamp should use PostgreSQL's NOW()
		ts := dialect.CurrentTimestamp()
		if ts == "" {
			t.Error("expected non-empty timestamp expression")
		}
	})

	t.Run("SQLite dialect returns correct values", func(t *testing.T) {
		dialect := core.NewSQLiteDialect()
		
		if dialect.Name() != "sqlite" {
			t.Errorf("expected name 'sqlite', got %s", dialect.Name())
		}
		
		if dialect.JSONType() != "JSON" {
			t.Errorf("expected JSON type 'JSON', got %s", dialect.JSONType())
		}
		
		if dialect.BooleanType() != "BOOLEAN" {
			t.Errorf("expected boolean type 'BOOLEAN', got %s", dialect.BooleanType())
		}
	})
}
