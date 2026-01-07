package dbmigrate_test

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/dbmigrate"
)

func TestSchemaExporter(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	t.Run("ExportCollectionsSchema for SQLite", func(t *testing.T) {
		dialect := core.NewSQLiteDialect()
		exporter := dbmigrate.NewSchemaExporter(app, dialect)

		schema, err := exporter.ExportCollectionsSchema()
		if err != nil {
			t.Fatalf("ExportCollectionsSchema failed: %v", err)
		}

		// Check that schema contains expected elements
		if !strings.Contains(schema, "CREATE TABLE") {
			t.Error("schema should contain CREATE TABLE statements")
		}

		if !strings.Contains(schema, "_params") {
			t.Error("schema should contain _params table")
		}

		if !strings.Contains(schema, "_collections") {
			t.Error("schema should contain _collections table")
		}

		if !strings.Contains(schema, "_logs") {
			t.Error("schema should contain _logs table")
		}

		// SQLite-specific: should contain SQLite-style random ID
		if !strings.Contains(schema, "randomblob") {
			t.Error("SQLite schema should contain randomblob")
		}
	})

	t.Run("ExportCollectionsSchema for PostgreSQL", func(t *testing.T) {
		dialect := core.NewPostgreSQLDialect()
		exporter := dbmigrate.NewSchemaExporter(app, dialect)

		schema, err := exporter.ExportCollectionsSchema()
		if err != nil {
			t.Fatalf("ExportCollectionsSchema failed: %v", err)
		}

		// Check that schema contains expected elements
		if !strings.Contains(schema, "CREATE TABLE") {
			t.Error("schema should contain CREATE TABLE statements")
		}

		// PostgreSQL-specific: should contain gen_random_bytes
		if !strings.Contains(schema, "gen_random_bytes") {
			t.Error("PostgreSQL schema should contain gen_random_bytes")
		}

		// PostgreSQL-specific: should use JSONB
		if !strings.Contains(schema, "JSONB") {
			t.Error("PostgreSQL schema should use JSONB type")
		}

		// PostgreSQL-specific: should use to_char for timestamps
		if !strings.Contains(schema, "to_char") {
			t.Error("PostgreSQL schema should contain to_char")
		}

		// PostgreSQL-specific: date_trunc for log index
		if !strings.Contains(schema, "date_trunc") {
			t.Error("PostgreSQL schema should contain date_trunc for hour index")
		}
	})
}

func TestPlanMigration(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	t.Run("creates migration plan for SQLite", func(t *testing.T) {
		dialect := core.NewSQLiteDialect()
		plan, err := dbmigrate.PlanMigration(app, dialect)
		if err != nil {
			t.Fatalf("PlanMigration failed: %v", err)
		}

		if plan.SchemaSQL == "" {
			t.Error("plan should have schema SQL")
		}

		// SQLite shouldn't have warnings
		if len(plan.Warnings) > 0 {
			t.Logf("SQLite warnings: %v", plan.Warnings)
		}
	})

	t.Run("creates migration plan for PostgreSQL with warnings", func(t *testing.T) {
		dialect := core.NewPostgreSQLDialect()
		plan, err := dbmigrate.PlanMigration(app, dialect)
		if err != nil {
			t.Fatalf("PlanMigration failed: %v", err)
		}

		if plan.SchemaSQL == "" {
			t.Error("plan should have schema SQL")
		}

		// PostgreSQL should have warnings
		if len(plan.Warnings) == 0 {
			t.Error("PostgreSQL plan should include warnings")
		}

		// Check for specific warning about experimental status
		hasExperimentalWarning := false
		for _, w := range plan.Warnings {
			if strings.Contains(w, "experimental") {
				hasExperimentalWarning = true
				break
			}
		}
		if !hasExperimentalWarning {
			t.Error("PostgreSQL plan should warn about experimental status")
		}
	})
}

func TestExportSchemaForDialect(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	t.Run("exports for SQLite type", func(t *testing.T) {
		schema, err := dbmigrate.ExportSchemaForDialect(app, core.DatabaseTypeSQLite)
		if err != nil {
			t.Fatalf("ExportSchemaForDialect failed: %v", err)
		}

		if schema == "" {
			t.Error("schema should not be empty")
		}

		if !strings.Contains(schema, "randomblob") {
			t.Error("SQLite schema should contain randomblob")
		}
	})

	t.Run("exports for PostgreSQL type", func(t *testing.T) {
		schema, err := dbmigrate.ExportSchemaForDialect(app, core.DatabaseTypePostgreSQL)
		if err != nil {
			t.Fatalf("ExportSchemaForDialect failed: %v", err)
		}

		if schema == "" {
			t.Error("schema should not be empty")
		}

		if !strings.Contains(schema, "gen_random_bytes") {
			t.Error("PostgreSQL schema should contain gen_random_bytes")
		}
	})
}
