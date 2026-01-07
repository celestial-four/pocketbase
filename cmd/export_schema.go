package cmd

import (
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/dbmigrate"
	"github.com/spf13/cobra"
)

// NewExportSchemaCommand creates and returns a new command for exporting
// the database schema in different formats (SQLite, PostgreSQL).
func NewExportSchemaCommand(app core.App) *cobra.Command {
	var format string
	var output string
	var includeData bool

	command := &cobra.Command{
		Use:   "export-schema",
		Short: "Export the database schema to a specific format",
		Long: `Export the PocketBase database schema to a specific format.

Supported formats:
  - sqlite (default): SQLite-compatible SQL
  - postgres: PostgreSQL-compatible SQL (experimental)

Examples:
  pocketbase export-schema                      # Export SQLite schema to stdout
  pocketbase export-schema --format postgres    # Export PostgreSQL schema to stdout
  pocketbase export-schema -o schema.sql        # Export to a file
  pocketbase export-schema --data               # Include data INSERT statements

Note: PostgreSQL support is experimental and not production-ready.`,
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			// Determine the target database type
			var dbType core.DatabaseType
			switch format {
			case "sqlite":
				dbType = core.DatabaseTypeSQLite
			case "postgres", "postgresql":
				dbType = core.DatabaseTypePostgreSQL
			default:
				return fmt.Errorf("unsupported format: %s (supported: sqlite, postgres)", format)
			}

			dialect := core.GetDialect(dbType)
			exporter := dbmigrate.NewSchemaExporter(app, dialect)

			// Export schema
			schemaSQL, err := exporter.ExportCollectionsSchema()
			if err != nil {
				return fmt.Errorf("failed to export schema: %w", err)
			}

			var result string
			result = schemaSQL

			// Include data if requested
			if includeData {
				dataSQL, err := exporter.ExportDataSQL()
				if err != nil {
					return fmt.Errorf("failed to export data: %w", err)
				}
				result += "\n" + dataSQL
			}

			// Output to file or stdout
			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
				fmt.Printf("Schema exported to %s\n", output)
			} else {
				fmt.Print(result)
			}

			// Print warnings for PostgreSQL
			if dbType == core.DatabaseTypePostgreSQL {
				fmt.Fprintln(os.Stderr, "\nWARNING: PostgreSQL support is experimental and not production-ready.")
				fmt.Fprintln(os.Stderr, "Please thoroughly test in a non-production environment before use.")
			}

			return nil
		},
	}

	command.Flags().StringVarP(
		&format,
		"format",
		"f",
		"sqlite",
		"Output format (sqlite or postgres)",
	)

	command.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"Output file path (default: stdout)",
	)

	command.Flags().BoolVar(
		&includeData,
		"data",
		false,
		"Include data INSERT statements",
	)

	return command
}

// NewMigrationPlanCommand creates a command for generating a migration plan
// from the current database to a target database type.
func NewMigrationPlanCommand(app core.App) *cobra.Command {
	var targetFormat string
	var output string

	command := &cobra.Command{
		Use:   "migration-plan",
		Short: "Generate a migration plan from current database to target format",
		Long: `Generate a migration plan for moving from the current database to a target format.

This command analyzes the current database schema and data, and generates
a migration plan including:
  - Schema SQL for the target database
  - Data INSERT statements
  - Warnings and potential issues

Supported targets:
  - postgres: PostgreSQL (experimental)

Examples:
  pocketbase migration-plan --target postgres           # Print plan to stdout
  pocketbase migration-plan --target postgres -o plan/  # Write to directory

Note: PostgreSQL support is experimental and not production-ready.`,
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			var dbType core.DatabaseType
			switch targetFormat {
			case "postgres", "postgresql":
				dbType = core.DatabaseTypePostgreSQL
			default:
				return fmt.Errorf("unsupported target: %s (supported: postgres)", targetFormat)
			}

			dialect := core.GetDialect(dbType)
			plan, err := dbmigrate.PlanMigration(app, dialect)
			if err != nil {
				return fmt.Errorf("failed to create migration plan: %w", err)
			}

			// Output results
			if output != "" {
				// Create output directory
				if err := os.MkdirAll(output, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}

				// Write schema
				schemaPath := output + "/schema.sql"
				if err := os.WriteFile(schemaPath, []byte(plan.SchemaSQL), 0644); err != nil {
					return fmt.Errorf("failed to write schema: %w", err)
				}
				fmt.Printf("Schema written to %s\n", schemaPath)

				// Write data
				dataPath := output + "/data.sql"
				if err := os.WriteFile(dataPath, []byte(plan.DataSQL), 0644); err != nil {
					return fmt.Errorf("failed to write data: %w", err)
				}
				fmt.Printf("Data written to %s\n", dataPath)

				// Write warnings
				if len(plan.Warnings) > 0 {
					warningsContent := "# Migration Warnings\n\n"
					for _, w := range plan.Warnings {
						warningsContent += "- " + w + "\n"
					}
					warningsPath := output + "/WARNINGS.md"
					if err := os.WriteFile(warningsPath, []byte(warningsContent), 0644); err != nil {
						return fmt.Errorf("failed to write warnings: %w", err)
					}
					fmt.Printf("Warnings written to %s\n", warningsPath)
				}
			} else {
				// Print to stdout
				fmt.Println("=== SCHEMA SQL ===")
				fmt.Println(plan.SchemaSQL)
				fmt.Println("\n=== DATA SQL ===")
				fmt.Println(plan.DataSQL)

				if len(plan.Warnings) > 0 {
					fmt.Println("\n=== WARNINGS ===")
					for _, w := range plan.Warnings {
						fmt.Println("- " + w)
					}
				}
			}

			return nil
		},
	}

	command.Flags().StringVarP(
		&targetFormat,
		"target",
		"t",
		"postgres",
		"Target database format (postgres)",
	)

	command.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"Output directory for migration files (default: stdout)",
	)

	return command
}
