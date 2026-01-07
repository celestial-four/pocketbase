package core_test

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestSQLiteDialect(t *testing.T) {
	d := core.NewSQLiteDialect()

	t.Run("Name", func(t *testing.T) {
		if d.Name() != "sqlite" {
			t.Errorf("expected 'sqlite', got '%s'", d.Name())
		}
	})

	t.Run("QuoteIdentifier", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"column", "`column`"},
			{"table_name", "`table_name`"},
			{"with`backtick", "`with``backtick`"},
		}

		for _, tt := range tests {
			result := d.QuoteIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("RandomID", func(t *testing.T) {
		result := d.RandomID()
		if !strings.Contains(result, "randomblob") {
			t.Errorf("RandomID() should contain 'randomblob', got %q", result)
		}
		if !strings.Contains(result, "hex") {
			t.Errorf("RandomID() should contain 'hex', got %q", result)
		}
	})

	t.Run("CurrentTimestamp", func(t *testing.T) {
		result := d.CurrentTimestamp()
		if !strings.Contains(result, "strftime") {
			t.Errorf("CurrentTimestamp() should contain 'strftime', got %q", result)
		}
	})

	t.Run("BooleanType", func(t *testing.T) {
		if d.BooleanType() != "BOOLEAN" {
			t.Errorf("expected 'BOOLEAN', got '%s'", d.BooleanType())
		}
	})

	t.Run("JSONType", func(t *testing.T) {
		if d.JSONType() != "JSON" {
			t.Errorf("expected 'JSON', got '%s'", d.JSONType())
		}
	})

	t.Run("AutoIncrementPK", func(t *testing.T) {
		result := d.AutoIncrementPK(false)
		if !strings.Contains(result, "AUTOINCREMENT") {
			t.Errorf("AutoIncrementPK should contain 'AUTOINCREMENT', got %q", result)
		}
	})

	t.Run("Placeholder", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			result := d.Placeholder(i)
			if result != "?" {
				t.Errorf("Placeholder(%d) = %q, want '?'", i, result)
			}
		}
	})

	t.Run("StrftimeHour", func(t *testing.T) {
		result := d.StrftimeHour("created")
		if !strings.Contains(result, "strftime") {
			t.Errorf("StrftimeHour() should contain 'strftime', got %q", result)
		}
		if !strings.Contains(result, "created") {
			t.Errorf("StrftimeHour() should contain column name, got %q", result)
		}
	})
}

func TestPostgreSQLDialect(t *testing.T) {
	d := core.NewPostgreSQLDialect()

	t.Run("Name", func(t *testing.T) {
		if d.Name() != "postgres" {
			t.Errorf("expected 'postgres', got '%s'", d.Name())
		}
	})

	t.Run("QuoteIdentifier", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"column", "\"column\""},
			{"table_name", "\"table_name\""},
			{"with\"quote", "\"with\"\"quote\""},
		}

		for _, tt := range tests {
			result := d.QuoteIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("RandomID", func(t *testing.T) {
		result := d.RandomID()
		if !strings.Contains(result, "gen_random_bytes") {
			t.Errorf("RandomID() should contain 'gen_random_bytes', got %q", result)
		}
		if !strings.Contains(result, "encode") {
			t.Errorf("RandomID() should contain 'encode', got %q", result)
		}
	})

	t.Run("CurrentTimestamp", func(t *testing.T) {
		result := d.CurrentTimestamp()
		if !strings.Contains(result, "to_char") {
			t.Errorf("CurrentTimestamp() should contain 'to_char', got %q", result)
		}
		if !strings.Contains(result, "now()") {
			t.Errorf("CurrentTimestamp() should contain 'now()', got %q", result)
		}
	})

	t.Run("BooleanType", func(t *testing.T) {
		if d.BooleanType() != "BOOLEAN" {
			t.Errorf("expected 'BOOLEAN', got '%s'", d.BooleanType())
		}
	})

	t.Run("JSONType", func(t *testing.T) {
		if d.JSONType() != "JSONB" {
			t.Errorf("expected 'JSONB', got '%s'", d.JSONType())
		}
	})

	t.Run("AutoIncrementPK", func(t *testing.T) {
		result := d.AutoIncrementPK(false)
		if result != "SERIAL PRIMARY KEY" {
			t.Errorf("AutoIncrementPK(false) = %q, want 'SERIAL PRIMARY KEY'", result)
		}

		result = d.AutoIncrementPK(true)
		if result != "BIGSERIAL PRIMARY KEY" {
			t.Errorf("AutoIncrementPK(true) = %q, want 'BIGSERIAL PRIMARY KEY'", result)
		}
	})

	t.Run("Placeholder", func(t *testing.T) {
		tests := []struct {
			index    int
			expected string
		}{
			{1, "$1"},
			{2, "$2"},
			{10, "$10"},
		}

		for _, tt := range tests {
			result := d.Placeholder(tt.index)
			if result != tt.expected {
				t.Errorf("Placeholder(%d) = %q, want %q", tt.index, result, tt.expected)
			}
		}
	})

	t.Run("StrftimeHour", func(t *testing.T) {
		result := d.StrftimeHour("created")
		if !strings.Contains(result, "date_trunc") {
			t.Errorf("StrftimeHour() should contain 'date_trunc', got %q", result)
		}
		if !strings.Contains(result, "hour") {
			t.Errorf("StrftimeHour() should contain 'hour', got %q", result)
		}
	})
}

func TestPostgreSQLDialectTranslateSQL(t *testing.T) {
	d := core.NewPostgreSQLDialect()

	t.Run("translates hex(randomblob)", func(t *testing.T) {
		sql := "SELECT hex(randomblob(7)) as id"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "gen_random_bytes") {
			t.Errorf("should translate randomblob to gen_random_bytes, got %q", result)
		}
		if !strings.Contains(result, "encode") {
			t.Errorf("should translate hex to encode, got %q", result)
		}
	})

	t.Run("translates strftime", func(t *testing.T) {
		sql := "SELECT strftime('%Y-%m-%d', created) FROM table"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "to_char") {
			t.Errorf("should translate strftime to to_char, got %q", result)
		}
	})

	t.Run("translates json_array", func(t *testing.T) {
		sql := "SELECT json_array(a, b, c)"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "jsonb_build_array") {
			t.Errorf("should translate json_array to jsonb_build_array, got %q", result)
		}
	})

	t.Run("translates json_object", func(t *testing.T) {
		sql := "SELECT json_object('key', value)"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "jsonb_build_object") {
			t.Errorf("should translate json_object to jsonb_build_object, got %q", result)
		}
	})

	t.Run("translates json_each", func(t *testing.T) {
		sql := "SELECT * FROM json_each(data)"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "jsonb_array_elements") {
			t.Errorf("should translate json_each to jsonb_array_elements, got %q", result)
		}
	})

	t.Run("translates json_array_length", func(t *testing.T) {
		sql := "SELECT json_array_length(data)"
		result := d.TranslateSQL(sql)
		if !strings.Contains(result, "jsonb_array_length") {
			t.Errorf("should translate json_array_length to jsonb_array_length, got %q", result)
		}
	})
}

func TestGetDialect(t *testing.T) {
	t.Run("returns SQLite dialect for SQLite type", func(t *testing.T) {
		d := core.GetDialect(core.DatabaseTypeSQLite)
		if d.Name() != "sqlite" {
			t.Errorf("expected sqlite dialect, got %s", d.Name())
		}
	})

	t.Run("returns PostgreSQL dialect for PostgreSQL type", func(t *testing.T) {
		d := core.GetDialect(core.DatabaseTypePostgreSQL)
		if d.Name() != "postgres" {
			t.Errorf("expected postgres dialect, got %s", d.Name())
		}
	})

	t.Run("returns SQLite dialect for empty type", func(t *testing.T) {
		d := core.GetDialect("")
		if d.Name() != "sqlite" {
			t.Errorf("expected sqlite dialect as default, got %s", d.Name())
		}
	})

	t.Run("returns SQLite dialect for unknown type", func(t *testing.T) {
		d := core.GetDialect("mysql")
		if d.Name() != "sqlite" {
			t.Errorf("expected sqlite dialect for unknown type, got %s", d.Name())
		}
	})
}

func TestSQLTranslator(t *testing.T) {
	t.Run("same dialect returns unchanged SQL", func(t *testing.T) {
		translator := core.NewSQLTranslator(core.NewSQLiteDialect(), core.NewSQLiteDialect())
		sql := "SELECT * FROM users"
		result := translator.Translate(sql)
		if result != sql {
			t.Errorf("same dialect should return unchanged SQL, got %q", result)
		}
	})

	t.Run("SQLite to PostgreSQL translation", func(t *testing.T) {
		translator := core.NewSQLTranslator(core.NewSQLiteDialect(), core.NewPostgreSQLDialect())
		sql := "SELECT hex(randomblob(7))"
		result := translator.Translate(sql)
		if !strings.Contains(result, "gen_random_bytes") {
			t.Errorf("should translate SQLite to PostgreSQL, got %q", result)
		}
	})
}

func TestTranslateSQLiteToPostgres(t *testing.T) {
	t.Run("translates complex SQL", func(t *testing.T) {
		sql := `
			CREATE TABLE users (
				id TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
				created TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL
			)
		`
		result := core.TranslateSQLiteToPostgres(sql)

		if !strings.Contains(result, "gen_random_bytes") {
			t.Errorf("should translate randomblob, got %q", result)
		}
		if !strings.Contains(result, "to_char") {
			t.Errorf("should translate strftime, got %q", result)
		}
	})
}

func TestTranslateCreateTable(t *testing.T) {
	sqliteDialect := core.NewSQLiteDialect()
	pgDialect := core.NewPostgreSQLDialect()

	t.Run("SQLite dialect returns unchanged", func(t *testing.T) {
		sql := "CREATE TABLE test (id TEXT PRIMARY KEY)"
		result := core.TranslateCreateTable(sql, sqliteDialect)
		if result != sql {
			t.Errorf("SQLite dialect should return unchanged SQL, got %q", result)
		}
	})

	t.Run("PostgreSQL translation", func(t *testing.T) {
		sql := `CREATE TABLE users (
			id TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
			data JSON DEFAULT "{}" NOT NULL
		)`
		result := core.TranslateCreateTable(sql, pgDialect)

		if !strings.Contains(result, "gen_random_bytes") {
			t.Errorf("should translate randomblob, got %q", result)
		}
		if !strings.Contains(result, "JSONB") {
			t.Errorf("should translate JSON to JSONB, got %q", result)
		}
	})
}
