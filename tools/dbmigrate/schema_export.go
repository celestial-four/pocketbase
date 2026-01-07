// Package dbmigrate provides utilities for database migration and schema export.
//
// This package offers tools for:
//   - Exporting PocketBase schemas to different database formats
//   - Migrating data from SQLite to PostgreSQL (and vice versa in the future)
//   - Converting SQLite-specific SQL to PostgreSQL-compatible SQL
//
// Note: PostgreSQL support in PocketBase is experimental. This package provides
// the foundation for future full PostgreSQL support.
package dbmigrate

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// SchemaExporter exports PocketBase database schemas to different formats.
type SchemaExporter struct {
	app     core.App
	dialect core.Dialect
}

// NewSchemaExporter creates a new schema exporter for the given app and target dialect.
func NewSchemaExporter(app core.App, targetDialect core.Dialect) *SchemaExporter {
	return &SchemaExporter{
		app:     app,
		dialect: targetDialect,
	}
}

// ExportCollectionsSchema exports all collection table schemas.
// Returns the CREATE TABLE statements for all collections.
func (e *SchemaExporter) ExportCollectionsSchema() (string, error) {
	collections, err := e.app.FindAllCollections()
	if err != nil {
		return "", fmt.Errorf("failed to fetch collections: %w", err)
	}

	var result strings.Builder

	// Add header comment
	result.WriteString(fmt.Sprintf("-- PocketBase Schema Export for %s\n", e.dialect.Name()))
	result.WriteString("-- Generated schema from PocketBase collections\n")
	result.WriteString("-- WARNING: PostgreSQL support is experimental\n\n")

	// Export system tables first
	systemSQL, err := e.exportSystemTables()
	if err != nil {
		return "", fmt.Errorf("failed to export system tables: %w", err)
	}
	result.WriteString(systemSQL)
	result.WriteString("\n")

	// Export collection tables
	for _, collection := range collections {
		collSQL := e.exportCollectionTable(collection)
		result.WriteString(collSQL)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// exportSystemTables exports the core system tables (_params, _collections, etc.)
func (e *SchemaExporter) exportSystemTables() (string, error) {
	var result strings.Builder

	// _params table
	paramsSQL := e.generateParamsTable()
	result.WriteString(paramsSQL)
	result.WriteString("\n")

	// _collections table
	collectionsSQL := e.generateCollectionsTable()
	result.WriteString(collectionsSQL)
	result.WriteString("\n")

	// _logs table (for auxiliary DB)
	logsSQL := e.generateLogsTable()
	result.WriteString(logsSQL)
	result.WriteString("\n")

	return result.String(), nil
}

// generateParamsTable generates the _params table schema
func (e *SchemaExporter) generateParamsTable() string {
	var sql strings.Builder

	sql.WriteString("-- _params table\n")
	sql.WriteString("CREATE TABLE IF NOT EXISTS _params (\n")
	sql.WriteString(fmt.Sprintf("    id TEXT PRIMARY KEY DEFAULT %s NOT NULL,\n", e.dialect.RandomID()))
	sql.WriteString(fmt.Sprintf("    value %s DEFAULT NULL,\n", e.dialect.JSONType()))
	sql.WriteString(fmt.Sprintf("    created TEXT DEFAULT %s NOT NULL,\n", e.dialect.CurrentTimestamp()))
	sql.WriteString(fmt.Sprintf("    updated TEXT DEFAULT %s NOT NULL\n", e.dialect.CurrentTimestamp()))
	sql.WriteString(");\n")

	return sql.String()
}

// generateCollectionsTable generates the _collections table schema
func (e *SchemaExporter) generateCollectionsTable() string {
	var sql strings.Builder

	sql.WriteString("-- _collections table\n")
	sql.WriteString("CREATE TABLE IF NOT EXISTS _collections (\n")
	sql.WriteString(fmt.Sprintf("    id TEXT PRIMARY KEY DEFAULT %s NOT NULL,\n", e.dialect.RandomID()))
	sql.WriteString(fmt.Sprintf("    system %s DEFAULT %s NOT NULL,\n", e.dialect.BooleanType(), e.dialect.BooleanFalse()))
	sql.WriteString("    type TEXT DEFAULT 'base' NOT NULL,\n")
	sql.WriteString("    name TEXT UNIQUE NOT NULL,\n")
	sql.WriteString(fmt.Sprintf("    fields %s DEFAULT '[]' NOT NULL,\n", e.dialect.JSONType()))
	sql.WriteString(fmt.Sprintf("    indexes %s DEFAULT '[]' NOT NULL,\n", e.dialect.JSONType()))
	sql.WriteString("    listRule TEXT DEFAULT NULL,\n")
	sql.WriteString("    viewRule TEXT DEFAULT NULL,\n")
	sql.WriteString("    createRule TEXT DEFAULT NULL,\n")
	sql.WriteString("    updateRule TEXT DEFAULT NULL,\n")
	sql.WriteString("    deleteRule TEXT DEFAULT NULL,\n")
	sql.WriteString(fmt.Sprintf("    options %s DEFAULT '{}' NOT NULL,\n", e.dialect.JSONType()))
	sql.WriteString(fmt.Sprintf("    created TEXT DEFAULT %s NOT NULL,\n", e.dialect.CurrentTimestamp()))
	sql.WriteString(fmt.Sprintf("    updated TEXT DEFAULT %s NOT NULL\n", e.dialect.CurrentTimestamp()))
	sql.WriteString(");\n")
	sql.WriteString("CREATE INDEX IF NOT EXISTS idx__collections_type ON _collections (type);\n")

	return sql.String()
}

// generateLogsTable generates the _logs table schema
func (e *SchemaExporter) generateLogsTable() string {
	var sql strings.Builder

	sql.WriteString("-- _logs table (auxiliary database)\n")
	sql.WriteString("CREATE TABLE IF NOT EXISTS _logs (\n")
	sql.WriteString(fmt.Sprintf("    id TEXT PRIMARY KEY DEFAULT %s NOT NULL,\n", e.dialect.RandomID()))
	sql.WriteString("    level INTEGER DEFAULT 0 NOT NULL,\n")
	sql.WriteString("    message TEXT DEFAULT '' NOT NULL,\n")
	sql.WriteString(fmt.Sprintf("    data %s DEFAULT '{}' NOT NULL,\n", e.dialect.JSONType()))
	sql.WriteString(fmt.Sprintf("    created TEXT DEFAULT %s NOT NULL\n", e.dialect.CurrentTimestamp()))
	sql.WriteString(");\n")
	sql.WriteString("CREATE INDEX IF NOT EXISTS idx_logs_level ON _logs (level);\n")
	sql.WriteString("CREATE INDEX IF NOT EXISTS idx_logs_message ON _logs (message);\n")
	sql.WriteString(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_logs_created_hour ON _logs (%s);\n", e.dialect.StrftimeHour("created")))

	return sql.String()
}

// exportCollectionTable generates the CREATE TABLE statement for a collection
func (e *SchemaExporter) exportCollectionTable(collection *core.Collection) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("-- Collection: %s\n", collection.Name))
	sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", e.dialect.QuoteIdentifier(collection.Name)))
	sql.WriteString(fmt.Sprintf("    id TEXT PRIMARY KEY DEFAULT %s NOT NULL", e.dialect.RandomID()))

	// Add fields
	for _, field := range collection.Fields {
		if field.GetName() == "id" {
			continue // Skip id field as it's already added
		}

		fieldSQL := e.generateFieldSQL(field)
		if fieldSQL != "" {
			sql.WriteString(",\n    ")
			sql.WriteString(fieldSQL)
		}
	}

	sql.WriteString("\n);\n")

	// Add indexes
	for _, index := range collection.Indexes {
		indexSQL := e.translateIndex(index, collection.Name)
		if indexSQL != "" {
			sql.WriteString(indexSQL)
			sql.WriteString("\n")
		}
	}

	return sql.String()
}

// generateFieldSQL generates the column definition for a field
func (e *SchemaExporter) generateFieldSQL(field core.Field) string {
	name := field.GetName()

	// Get the column type from the app's field definition
	// This handles all the database-specific types correctly
	columnType := field.ColumnType(e.app)

	// Translate SQLite column type to PostgreSQL if needed
	if e.dialect.Name() == "postgres" {
		columnType = translateColumnType(columnType, e.dialect)
	}

	var result strings.Builder
	result.WriteString(e.dialect.QuoteIdentifier(name))
	result.WriteString(" ")
	result.WriteString(columnType)

	return result.String()
}

// translateColumnType converts a SQLite column type definition to PostgreSQL
func translateColumnType(columnType string, dialect core.Dialect) string {
	result := columnType

	// Replace JSON with JSONB for PostgreSQL
	result = strings.ReplaceAll(result, " JSON ", " JSONB ")
	result = strings.ReplaceAll(result, " JSON\n", " JSONB\n")
	if strings.HasSuffix(result, " JSON") {
		result = result[:len(result)-5] + " JSONB"
	}

	return result
}

// translateIndex translates an index definition to the target dialect
func (e *SchemaExporter) translateIndex(indexDef string, tableName string) string {
	if e.dialect.Name() == "sqlite" {
		return indexDef // No translation needed for SQLite
	}

	// For PostgreSQL, we need to translate SQLite-specific expressions
	result := indexDef

	// Translate strftime in index expressions
	if strings.Contains(result, "strftime") {
		// Replace SQLite strftime with PostgreSQL date_trunc or to_char
		result = core.TranslateSQLiteToPostgres(result)
	}

	return result
}

// ExportDataSQL exports all data as INSERT statements.
// This is useful for migrating data between databases.
func (e *SchemaExporter) ExportDataSQL() (string, error) {
	collections, err := e.app.FindAllCollections()
	if err != nil {
		return "", fmt.Errorf("failed to fetch collections: %w", err)
	}

	var result strings.Builder

	result.WriteString("-- Data Export\n")
	result.WriteString("-- WARNING: This is a simplified export. Complex data types may need manual review.\n\n")

	for _, collection := range collections {
		dataSQL, err := e.exportCollectionData(collection)
		if err != nil {
			return "", fmt.Errorf("failed to export data for %s: %w", collection.Name, err)
		}
		result.WriteString(dataSQL)
	}

	return result.String(), nil
}

// exportCollectionData exports all records from a collection as INSERT statements
func (e *SchemaExporter) exportCollectionData(collection *core.Collection) (string, error) {
	records, err := e.app.FindAllRecords(collection.Name)
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("-- Data for collection: %s\n", collection.Name))

	for _, record := range records {
		insertSQL := e.generateInsertSQL(collection, record)
		result.WriteString(insertSQL)
		result.WriteString("\n")
	}
	result.WriteString("\n")

	return result.String(), nil
}

// generateInsertSQL generates an INSERT statement for a record
func (e *SchemaExporter) generateInsertSQL(collection *core.Collection, record *core.Record) string {
	var columns []string
	var values []string

	// Always include id
	columns = append(columns, e.dialect.QuoteIdentifier("id"))
	values = append(values, fmt.Sprintf("'%s'", escapeString(record.Id)))

	for _, field := range collection.Fields {
		if field.GetName() == "id" {
			continue
		}

		columns = append(columns, e.dialect.QuoteIdentifier(field.GetName()))
		value := e.formatValue(record.Get(field.GetName()), field.Type())
		values = append(values, value)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);",
		e.dialect.QuoteIdentifier(collection.Name),
		strings.Join(columns, ", "),
		strings.Join(values, ", "))
}

// formatValue formats a value for SQL insertion
// Note: This is a simplified implementation for the export tool.
// For production data migration, use parameterized queries instead.
func (e *SchemaExporter) formatValue(value any, fieldType string) string {
	if value == nil {
		return "NULL"
	}

	switch fieldType {
	case core.FieldTypeNumber:
		// Handle different numeric types explicitly
		switch v := value.(type) {
		case float64:
			return fmt.Sprintf("%g", v) // %g removes trailing zeros
		case float32:
			return fmt.Sprintf("%g", v)
		case int:
			return fmt.Sprintf("%d", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case int32:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", value)
		}
	case core.FieldTypeBool:
		if b, ok := value.(bool); ok && b {
			return e.dialect.BooleanTrue()
		}
		return e.dialect.BooleanFalse()
	case core.FieldTypeJSON, core.FieldTypeSelect, core.FieldTypeRelation, core.FieldTypeFile, core.FieldTypeGeoPoint:
		// JSON and array values need proper JSON marshaling
		jsonStr := formatJSONValue(value)
		return fmt.Sprintf("'%s'", escapeSQL(jsonStr))
	default:
		str := fmt.Sprintf("%v", value)
		return fmt.Sprintf("'%s'", escapeSQL(str))
	}
}

// formatJSONValue converts a value to a JSON string representation
func formatJSONValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		// For maps, slices, etc. - use simple string conversion
		// A full implementation would use json.Marshal
		return fmt.Sprintf("%v", v)
	}
}

// escapeSQL escapes a string for SQL insertion.
// This handles the most common cases: single quotes and backslashes.
// For more complex scenarios, consider using parameterized queries.
func escapeSQL(s string) string {
	// Escape backslashes first, then single quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "''")
	// Escape null bytes which can cause issues
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}

// Legacy function kept for backward compatibility
func escapeString(s string) string {
	return escapeSQL(s)
}

// MigrationPlan represents a plan for migrating from one database to another.
type MigrationPlan struct {
	SchemaSQL string   // SQL to create schema
	DataSQL   string   // SQL to insert data
	Warnings  []string // Any warnings or issues found
}

// PlanMigration creates a migration plan from the current database to the target dialect.
func PlanMigration(app core.App, targetDialect core.Dialect) (*MigrationPlan, error) {
	exporter := NewSchemaExporter(app, targetDialect)

	plan := &MigrationPlan{
		Warnings: make([]string, 0),
	}

	// Export schema
	schemaSQL, err := exporter.ExportCollectionsSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to export schema: %w", err)
	}
	plan.SchemaSQL = schemaSQL

	// Export data
	dataSQL, err := exporter.ExportDataSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to export data: %w", err)
	}
	plan.DataSQL = dataSQL

	// Add warnings for PostgreSQL
	if targetDialect.Name() == "postgres" {
		plan.Warnings = append(plan.Warnings,
			"PostgreSQL support is experimental and not production-ready",
			"File storage works differently in PostgreSQL - plan accordingly",
			"Some SQLite-specific features may not work correctly",
			"Thoroughly test the migration in a non-production environment first",
		)
	}

	return plan, nil
}

// ExportSchemaForDialect is a convenience function to export the schema
// for a specific database type.
func ExportSchemaForDialect(app core.App, dbType core.DatabaseType) (string, error) {
	dialect := core.GetDialect(dbType)
	exporter := NewSchemaExporter(app, dialect)
	return exporter.ExportCollectionsSchema()
}
