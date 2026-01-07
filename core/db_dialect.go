package core

import (
	"fmt"
	"regexp"
	"strings"
)

// Dialect defines the interface for database-specific SQL operations.
// This allows the same high-level operations to be translated into
// database-specific SQL syntax.
type Dialect interface {
	// Name returns the dialect identifier (e.g., "sqlite", "postgres").
	Name() string

	// QuoteIdentifier returns the properly quoted identifier for the dialect.
	// SQLite and PostgreSQL use different quoting conventions.
	QuoteIdentifier(identifier string) string

	// RandomID returns the SQL expression for generating a random ID.
	// SQLite: ('r'||lower(hex(randomblob(7))))
	// PostgreSQL: ('r' || encode(gen_random_bytes(7), 'hex'))
	RandomID() string

	// CurrentTimestamp returns the SQL expression for the current timestamp.
	// SQLite: strftime('%Y-%m-%d %H:%M:%fZ')
	// PostgreSQL: to_char(now() at time zone 'UTC', 'YYYY-MM-DD HH24:MI:SS.MS"Z"')
	CurrentTimestamp() string

	// BooleanType returns the SQL type name for boolean values.
	BooleanType() string

	// BooleanTrue returns the SQL literal for true.
	BooleanTrue() string

	// BooleanFalse returns the SQL literal for false.
	BooleanFalse() string

	// JSONType returns the SQL type name for JSON data.
	JSONType() string

	// AutoIncrementPK returns the SQL syntax for an auto-increment primary key.
	// SQLite: INTEGER PRIMARY KEY AUTOINCREMENT
	// PostgreSQL: SERIAL PRIMARY KEY or BIGSERIAL PRIMARY KEY
	AutoIncrementPK(big bool) string

	// Placeholder returns the parameter placeholder for the given index.
	// SQLite: ?
	// PostgreSQL: $1, $2, ...
	Placeholder(index int) string

	// StrftimeHour returns SQL to extract date truncated to hour.
	// Used for log indexing.
	StrftimeHour(column string) string

	// TranslateSQL translates generic SQL with dialect markers to this dialect.
	// This handles common patterns like [[column]], {{table}}, BOOLEAN, etc.
	TranslateSQL(sql string) string
}

// SQLiteDialect implements Dialect for SQLite databases.
type SQLiteDialect struct{}

// NewSQLiteDialect creates a new SQLite dialect instance.
func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{}
}

func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

func (d *SQLiteDialect) QuoteIdentifier(identifier string) string {
	// SQLite uses backticks or double quotes; we use backticks for consistency
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func (d *SQLiteDialect) RandomID() string {
	return "('r'||lower(hex(randomblob(7))))"
}

func (d *SQLiteDialect) CurrentTimestamp() string {
	return "(strftime('%Y-%m-%d %H:%M:%fZ'))"
}

func (d *SQLiteDialect) BooleanType() string {
	return "BOOLEAN"
}

func (d *SQLiteDialect) BooleanTrue() string {
	return "TRUE"
}

func (d *SQLiteDialect) BooleanFalse() string {
	return "FALSE"
}

func (d *SQLiteDialect) JSONType() string {
	return "JSON"
}

func (d *SQLiteDialect) AutoIncrementPK(big bool) string {
	return "INTEGER PRIMARY KEY AUTOINCREMENT"
}

func (d *SQLiteDialect) Placeholder(index int) string {
	return "?"
}

func (d *SQLiteDialect) StrftimeHour(column string) string {
	return fmt.Sprintf("strftime('%%Y-%%m-%%d %%H:00:00', %s)", column)
}

func (d *SQLiteDialect) TranslateSQL(sql string) string {
	// SQLite is the native dialect, no translation needed for basic markers
	return sql
}

// PostgreSQLDialect implements Dialect for PostgreSQL databases.
type PostgreSQLDialect struct{}

// NewPostgreSQLDialect creates a new PostgreSQL dialect instance.
func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{}
}

func (d *PostgreSQLDialect) Name() string {
	return "postgres"
}

func (d *PostgreSQLDialect) QuoteIdentifier(identifier string) string {
	// PostgreSQL uses double quotes for identifiers
	return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
}

func (d *PostgreSQLDialect) RandomID() string {
	// PostgreSQL uses gen_random_bytes() for random bytes and encode() for hex
	return "('r' || encode(gen_random_bytes(7), 'hex'))"
}

func (d *PostgreSQLDialect) CurrentTimestamp() string {
	// PostgreSQL timestamp with milliseconds in ISO format
	return "(to_char(now() at time zone 'UTC', 'YYYY-MM-DD HH24:MI:SS.MS') || 'Z')"
}

func (d *PostgreSQLDialect) BooleanType() string {
	return "BOOLEAN"
}

func (d *PostgreSQLDialect) BooleanTrue() string {
	return "TRUE"
}

func (d *PostgreSQLDialect) BooleanFalse() string {
	return "FALSE"
}

func (d *PostgreSQLDialect) JSONType() string {
	// Use JSONB for better performance and indexing capabilities
	return "JSONB"
}

func (d *PostgreSQLDialect) AutoIncrementPK(big bool) string {
	if big {
		return "BIGSERIAL PRIMARY KEY"
	}
	return "SERIAL PRIMARY KEY"
}

func (d *PostgreSQLDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d *PostgreSQLDialect) StrftimeHour(column string) string {
	return fmt.Sprintf("date_trunc('hour', %s::timestamp)", column)
}

// Common patterns for SQL translation
var (
	// Matches SQLite hex(randomblob(n)) pattern
	randomBlobPattern = regexp.MustCompile(`(?i)hex\s*\(\s*randomblob\s*\(\s*(\d+)\s*\)\s*\)`)

	// Matches SQLite strftime pattern
	strftimePattern = regexp.MustCompile(`(?i)strftime\s*\(\s*'([^']+)'\s*(?:,\s*([^)]+))?\s*\)`)

	// Matches SQLite iif function
	iifPattern = regexp.MustCompile(`(?i)\biif\s*\(`)

	// Matches json_valid function
	jsonValidPattern = regexp.MustCompile(`(?i)\bjson_valid\s*\(`)

	// Matches json_type function
	jsonTypePattern = regexp.MustCompile(`(?i)\bjson_type\s*\(`)

	// Matches json_array function
	jsonArrayPattern = regexp.MustCompile(`(?i)\bjson_array\s*\(`)

	// Matches JSON_EXTRACT function
	jsonExtractPattern = regexp.MustCompile(`(?i)\bJSON_EXTRACT\s*\(`)

	// Matches json_each function
	jsonEachPattern = regexp.MustCompile(`(?i)\bjson_each\s*\(`)

	// Matches json_array_length function
	jsonArrayLengthPattern = regexp.MustCompile(`(?i)\bjson_array_length\s*\(`)

	// Matches json_object function
	jsonObjectPattern = regexp.MustCompile(`(?i)\bjson_object\s*\(`)
)

func (d *PostgreSQLDialect) TranslateSQL(sql string) string {
	result := sql

	// Translate hex(randomblob(n)) to encode(gen_random_bytes(n), 'hex')
	result = randomBlobPattern.ReplaceAllString(result, "encode(gen_random_bytes($1), 'hex')")

	// Translate strftime patterns
	result = d.translateStrftime(result)

	// Translate iif(a,b,c) to CASE WHEN a THEN b ELSE c END
	// Note: This is a simplified replacement; complex nested iif would need proper parsing
	result = iifPattern.ReplaceAllString(result, "CASE WHEN (")

	// Translate SQLite JSON functions to PostgreSQL equivalents
	result = d.translateJSONFunctions(result)

	return result
}

// translateStrftime converts SQLite strftime calls to PostgreSQL to_char
func (d *PostgreSQLDialect) translateStrftime(sql string) string {
	return strftimePattern.ReplaceAllStringFunc(sql, func(match string) string {
		submatches := strftimePattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		format := submatches[1]
		column := "now() at time zone 'UTC'"
		if len(submatches) >= 3 && submatches[2] != "" {
			column = strings.TrimSpace(submatches[2])
			if column != "" {
				column = column + "::timestamp at time zone 'UTC'"
			} else {
				column = "now() at time zone 'UTC'"
			}
		}

		// Convert SQLite format to PostgreSQL format
		pgFormat := convertStrftimeFormat(format)

		return fmt.Sprintf("to_char(%s, '%s')", column, pgFormat)
	})
}

// convertStrftimeFormat converts SQLite strftime format to PostgreSQL to_char format
func convertStrftimeFormat(sqliteFormat string) string {
	// Map SQLite format specifiers to PostgreSQL
	replacements := map[string]string{
		"%Y": "YYYY",
		"%m": "MM",
		"%d": "DD",
		"%H": "HH24",
		"%M": "MI",
		"%S": "SS",
		"%f": "MS", // SQLite %f includes seconds with fraction, but we use MS for milliseconds
		"%j": "DDD",
		"%W": "IW",
		"%w": "D",
		"%%": "%",
	}

	result := sqliteFormat
	for sqlite, pg := range replacements {
		result = strings.ReplaceAll(result, sqlite, pg)
	}

	// Handle the 'Z' suffix specially - it should remain as a literal
	result = strings.ReplaceAll(result, "Z\"", "\"Z\"")

	return result
}

// translateJSONFunctions converts SQLite JSON functions to PostgreSQL equivalents
func (d *PostgreSQLDialect) translateJSONFunctions(sql string) string {
	result := sql

	// json_valid(x) -> (x)::jsonb IS NOT NULL (simplified)
	// Note: PostgreSQL validates JSON on cast, so we use a try-cast pattern
	result = jsonValidPattern.ReplaceAllString(result, "(")

	// json_type(x) -> jsonb_typeof(x::jsonb)
	result = jsonTypePattern.ReplaceAllString(result, "jsonb_typeof(")

	// json_array(x) -> jsonb_build_array(x)
	result = jsonArrayPattern.ReplaceAllString(result, "jsonb_build_array(")

	// JSON_EXTRACT(x, '$.path') -> x::jsonb #>> '{path}'
	// Note: This is simplified; complex paths need more handling
	result = jsonExtractPattern.ReplaceAllString(result, "jsonb_extract_path_text(")

	// json_each(x) -> jsonb_array_elements(x::jsonb)
	result = jsonEachPattern.ReplaceAllString(result, "jsonb_array_elements(")

	// json_array_length(x) -> jsonb_array_length(x::jsonb)
	result = jsonArrayLengthPattern.ReplaceAllString(result, "jsonb_array_length(")

	// json_object(k, v, ...) -> jsonb_build_object(k, v, ...)
	result = jsonObjectPattern.ReplaceAllString(result, "jsonb_build_object(")

	return result
}

// GetDialect returns the appropriate Dialect implementation for the given database type.
func GetDialect(dbType DatabaseType) Dialect {
	switch dbType {
	case DatabaseTypePostgreSQL:
		return NewPostgreSQLDialect()
	default:
		return NewSQLiteDialect()
	}
}

// SQLTranslator provides methods for translating SQL between dialects.
type SQLTranslator struct {
	fromDialect Dialect
	toDialect   Dialect
}

// NewSQLTranslator creates a new SQL translator.
func NewSQLTranslator(from, to Dialect) *SQLTranslator {
	return &SQLTranslator{
		fromDialect: from,
		toDialect:   to,
	}
}

// Translate converts SQL from the source dialect to the target dialect.
func (t *SQLTranslator) Translate(sql string) string {
	if t.fromDialect.Name() == t.toDialect.Name() {
		return sql
	}

	// For now, we only support translation from SQLite to PostgreSQL
	if t.fromDialect.Name() == "sqlite" && t.toDialect.Name() == "postgres" {
		return t.toDialect.TranslateSQL(sql)
	}

	// For other combinations, return the original SQL
	// (future work could add more translation paths)
	return sql
}

// TranslateSQLiteToPostgres is a convenience function for translating
// SQLite SQL to PostgreSQL SQL.
func TranslateSQLiteToPostgres(sql string) string {
	translator := NewSQLTranslator(NewSQLiteDialect(), NewPostgreSQLDialect())
	return translator.Translate(sql)
}

// TranslateCreateTable translates a CREATE TABLE statement from SQLite to PostgreSQL.
// This handles common data type and default value differences.
func TranslateCreateTable(sql string, toDialect Dialect) string {
	if toDialect.Name() == "sqlite" {
		return sql
	}

	result := sql

	// Replace SQLite-specific default ID generation
	sqliteIDDefault := `DEFAULT ('r'||lower(hex(randomblob(7))))`
	result = strings.ReplaceAll(result, sqliteIDDefault, fmt.Sprintf("DEFAULT %s", toDialect.RandomID()))

	// Replace strftime for created/updated timestamps
	sqliteStrftime := `DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ'))`
	result = strings.ReplaceAll(result, sqliteStrftime, fmt.Sprintf("DEFAULT %s", toDialect.CurrentTimestamp()))

	// Replace JSON type with JSONB for PostgreSQL
	result = strings.ReplaceAll(result, " JSON ", fmt.Sprintf(" %s ", toDialect.JSONType()))
	result = strings.ReplaceAll(result, " JSON\n", fmt.Sprintf(" %s\n", toDialect.JSONType()))

	// Translate the general SQL patterns
	result = toDialect.TranslateSQL(result)

	return result
}

// TranslateSQLForDB translates SQL from SQLite-syntax to the appropriate dialect
// based on the provided database type.
// This is a convenience function for use in migrations and raw SQL queries.
func TranslateSQLForDB(sql string, dbType DatabaseType) string {
	if dbType == DatabaseTypePostgreSQL {
		return TranslateCreateTable(sql, NewPostgreSQLDialect())
	}
	return sql
}
