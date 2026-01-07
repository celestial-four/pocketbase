package core_test

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestGetDatabaseConfigFromEnv(t *testing.T) {
	// Save original env vars
	origType := os.Getenv("PB_DB_TYPE")
	origURL := os.Getenv("PB_DB_URL")
	defer func() {
		os.Setenv("PB_DB_TYPE", origType)
		os.Setenv("PB_DB_URL", origURL)
	}()

	t.Run("defaults to SQLite", func(t *testing.T) {
		os.Unsetenv("PB_DB_TYPE")
		os.Unsetenv("PB_DB_URL")

		config := core.GetDatabaseConfigFromEnv()
		if !config.IsSQLite() {
			t.Errorf("expected SQLite as default, got %s", config.Type)
		}
		if config.Type != core.DatabaseTypeSQLite {
			t.Errorf("expected Type to be %s, got %s", core.DatabaseTypeSQLite, config.Type)
		}
	})

	t.Run("reads PB_DB_TYPE environment variable", func(t *testing.T) {
		os.Setenv("PB_DB_TYPE", "sqlite")
		os.Unsetenv("PB_DB_URL")

		config := core.GetDatabaseConfigFromEnv()
		if config.Type != core.DatabaseTypeSQLite {
			t.Errorf("expected Type to be %s, got %s", core.DatabaseTypeSQLite, config.Type)
		}
	})

	t.Run("case insensitive type", func(t *testing.T) {
		os.Setenv("PB_DB_TYPE", "SQLITE")
		os.Unsetenv("PB_DB_URL")

		config := core.GetDatabaseConfigFromEnv()
		if config.Type != core.DatabaseTypeSQLite {
			t.Errorf("expected Type to be %s (lowercased), got %s", core.DatabaseTypeSQLite, config.Type)
		}
	})

	t.Run("reads PB_DB_URL environment variable", func(t *testing.T) {
		os.Setenv("PB_DB_TYPE", "postgres")
		os.Setenv("PB_DB_URL", "postgres://user:pass@localhost:5432/testdb")

		config := core.GetDatabaseConfigFromEnv()
		if config.URL != "postgres://user:pass@localhost:5432/testdb" {
			t.Errorf("expected URL to be postgres://user:pass@localhost:5432/testdb, got %s", config.URL)
		}
	})
}

func TestDatabaseConfigValidate(t *testing.T) {
	t.Run("SQLite is always valid", func(t *testing.T) {
		config := &core.DatabaseConfig{
			Type: core.DatabaseTypeSQLite,
		}
		if err := config.Validate(); err != nil {
			t.Errorf("expected SQLite config to be valid, got error: %v", err)
		}
	})

	t.Run("PostgreSQL returns not implemented error", func(t *testing.T) {
		config := &core.DatabaseConfig{
			Type: core.DatabaseTypePostgreSQL,
			URL:  "postgres://user:pass@localhost:5432/testdb",
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected PostgreSQL to return error (not implemented), got nil")
		}
	})

	t.Run("unknown type returns error", func(t *testing.T) {
		config := &core.DatabaseConfig{
			Type: "mysql",
		}
		err := config.Validate()
		if err == nil {
			t.Error("expected unknown type to return error, got nil")
		}
	})
}

func TestDatabaseConfigHelpers(t *testing.T) {
	t.Run("IsSQLite", func(t *testing.T) {
		tests := []struct {
			name     string
			dbType   core.DatabaseType
			expected bool
		}{
			{"explicit sqlite", core.DatabaseTypeSQLite, true},
			{"empty type", "", true},
			{"postgres", core.DatabaseTypePostgreSQL, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &core.DatabaseConfig{Type: tt.dbType}
				if got := config.IsSQLite(); got != tt.expected {
					t.Errorf("IsSQLite() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("IsPostgreSQL", func(t *testing.T) {
		tests := []struct {
			name     string
			dbType   core.DatabaseType
			expected bool
		}{
			{"postgres", core.DatabaseTypePostgreSQL, true},
			{"sqlite", core.DatabaseTypeSQLite, false},
			{"empty", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &core.DatabaseConfig{Type: tt.dbType}
				if got := config.IsPostgreSQL(); got != tt.expected {
					t.Errorf("IsPostgreSQL() = %v, want %v", got, tt.expected)
				}
			})
		}
	})
}

func TestBaseAppWithDatabaseConfig(t *testing.T) {
	const testDataDir = "./pb_base_app_dbconfig_test_data_dir/"
	defer os.RemoveAll(testDataDir)

	t.Run("creates app with default SQLite config", func(t *testing.T) {
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir,
			IsDev:   true,
		})

		if app == nil {
			t.Fatal("expected app to be created")
		}

		// App should have been created successfully with SQLite default
	})

	t.Run("creates app with explicit SQLite config", func(t *testing.T) {
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir,
			IsDev:   true,
			DBConfig: &core.DatabaseConfig{
				Type: core.DatabaseTypeSQLite,
			},
		})

		if app == nil {
			t.Fatal("expected app to be created")
		}
	})

	t.Run("falls back to SQLite with invalid config", func(t *testing.T) {
		// This should not panic, but gracefully fall back to SQLite
		app := core.NewBaseApp(core.BaseAppConfig{
			DataDir: testDataDir,
			IsDev:   true,
			DBConfig: &core.DatabaseConfig{
				Type: "invalid",
			},
		})

		if app == nil {
			t.Fatal("expected app to be created even with invalid config")
		}
	})
}
