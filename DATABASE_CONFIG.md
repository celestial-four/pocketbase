# Database Configuration

PocketBase includes infrastructure for supporting multiple database backends.

## Current Status

**SQLite Only**: Currently, only SQLite is fully supported and production-ready. PostgreSQL support is not yet implemented.

## Configuration Options

### Environment Variables

- `PB_DB_TYPE`: Database type (`sqlite` or `postgres`). Defaults to `sqlite`.
- `PB_DB_URL`: Database connection URL (required for `postgres`, ignored for `sqlite`).

### CLI Flags

- `--dbType`: Database type (`sqlite` or `postgres`). Defaults to `sqlite`.
- `--dbURL`: Database connection URL (required for `postgres`, ignored for `sqlite`).

### Priority

Configuration is resolved in the following order:
1. CLI flags (highest priority)
2. Environment variables
3. Config struct defaults
4. System defaults (SQLite)

## Usage

### Using SQLite (Default)

No configuration needed - SQLite is the default:

```bash
pocketbase serve
```

### Using Environment Variables

```bash
export PB_DB_TYPE=sqlite
pocketbase serve
```

### Using CLI Flags

```bash
pocketbase serve --dbType sqlite
```

## API Usage

When using PocketBase as a Go library, you can specify database configuration programmatically:

```go
package main

import (
    "log"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func main() {
    app := pocketbase.NewWithConfig(pocketbase.Config{
        // other config options...
    })

    // The database configuration is automatically loaded from environment variables
    // or you can set it explicitly in the BaseAppConfig

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## PostgreSQL Support (Future)

PostgreSQL support is planned but not yet implemented. The infrastructure exists to support it, including:

- Database configuration abstraction (`DatabaseConfig`)
- Connection management hooks
- CLI and environment variable support

Attempting to use PostgreSQL will result in a graceful fallback to SQLite with a warning message.

### What's Needed for PostgreSQL Support

Full PostgreSQL support requires:

1. **Schema Translation**: Convert SQLite-specific SQL to PostgreSQL
2. **Data Type Mapping**: Handle differences in data types
3. **Migration System**: Create PostgreSQL-compatible migrations
4. **Query Builder Updates**: Ensure dbx queries work with PostgreSQL
5. **Transaction Handling**: Test and verify PostgreSQL transaction semantics
6. **Testing**: Comprehensive test suite running against both databases
7. **Documentation**: Migration guide from SQLite to PostgreSQL

## Architecture

### DatabaseConfig

The `core.DatabaseConfig` struct provides a clean abstraction for database configuration:

```go
type DatabaseConfig struct {
    Type DatabaseType // "sqlite" or "postgres"
    URL  string       // Connection URL (for postgres)
}
```

### Methods

- `GetDatabaseConfigFromEnv()`: Reads configuration from environment variables
- `Validate()`: Validates the configuration
- `IsSQLite()`: Returns true if using SQLite
- `IsPostgreSQL()`: Returns true if using PostgreSQL

## Backward Compatibility

This change is 100% backward compatible:

- Existing applications continue to work without any changes
- SQLite remains the default database
- No breaking changes to the API
- All existing flags and configuration options work as before

## Production Readiness

The database configuration infrastructure is production-ready, but PostgreSQL support is not. For production use:

- ✅ Use SQLite (default, fully tested)
- ❌ Do not use PostgreSQL (not implemented)
