# PostgreSQL Support Status

## Current Status: Experimental (v1.0-beta)

PostgreSQL support in PocketBase is **experimental** and should be thoroughly tested before production use.

## What's Implemented ✅

### 1. Database Configuration Infrastructure
- **DatabaseConfig struct** (`core/db_config.go`) - type-safe configuration for database selection
- **Environment variables**: `PB_DB_TYPE` (sqlite/postgres) and `PB_DB_URL` (connection string)
- **CLI flags**: `--dbType` and `--dbURL`
- Configuration priority: CLI flags > Environment variables > Config defaults > SQLite

### 2. SQL Dialect Abstraction Layer (`core/db_dialect.go`)
- `Dialect` interface for database-specific SQL operations
- `SQLiteDialect` - SQLite-specific implementations
- `PostgreSQLDialect` - PostgreSQL-specific implementations
- Key translations:
  - `hex(randomblob(n))` → `encode(gen_random_bytes(n), 'hex')`
  - `strftime()` → `to_char()` with format conversion
  - `JSON` → `JSONB`
  - `json_array()` → `jsonb_build_array()`
  - `json_object()` → `jsonb_build_object()`
  - `json_each()` → `jsonb_array_elements()`
  - And more JSON function translations

### 3. PostgreSQL Connection (`core/db_connect_postgres.go`)
- Uses `pgx/v5` driver (high-performance PostgreSQL driver)
- Connection pooling inherited from `database/sql`
- SSL/TLS support via connection string parameters

### 4. Database-Specific Operations (`core/db_table.go`)
- `TableColumns()` - uses `information_schema` for PostgreSQL
- `TableInfo()` - uses `information_schema` for PostgreSQL  
- `TableIndexes()` - uses `pg_indexes` for PostgreSQL
- `HasTable()` - uses `information_schema` for PostgreSQL

### 5. Schema Export Tools (`tools/dbmigrate/`)
- Export PocketBase schemas to PostgreSQL format
- Generate migration plans
- Data export as INSERT statements
- CLI command: `pocketbase export-schema --format postgres`

### 6. Migrations
- System migrations in `migrations/` use `TranslateSQLForDB()` for compatibility
- Collections table creation works on both databases
- Logs table (auxiliary DB) has database-specific index handling

## How to Use PostgreSQL

### Option 1: Environment Variables
```bash
export PB_DB_TYPE=postgres
export PB_DB_URL="postgres://user:password@localhost:5432/pocketbase?sslmode=disable"
./pocketbase serve
```

### Option 2: CLI Flags
```bash
./pocketbase serve \
  --dbType postgres \
  --dbURL "postgres://user:password@localhost:5432/pocketbase?sslmode=disable"
```

### Option 3: Programmatic Configuration
```go
package main

import (
    "log"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func main() {
    app := core.NewBaseApp(core.BaseAppConfig{
        DataDir: "pb_data", // Still needed for file storage
        DBConfig: &core.DatabaseConfig{
            Type: core.DatabaseTypePostgreSQL,
            URL:  "postgres://user:password@localhost:5432/pocketbase?sslmode=disable",
        },
    })
    
    if err := app.Bootstrap(); err != nil {
        log.Fatal(err)
    }
    // ... rest of your application
}
```

### PostgreSQL Connection String Formats
```
# Standard format
postgres://user:password@host:port/database?sslmode=disable

# With SSL
postgres://user:password@host:port/database?sslmode=require

# Key-value format  
host=localhost port=5432 user=user password=password dbname=pocketbase sslmode=disable
```

## Migration from SQLite to PostgreSQL

### 1. Export Schema
```bash
./pocketbase export-schema --format postgres -o schema.sql
```

### 2. Export Data (Optional)
```bash
./pocketbase export-schema --format postgres --data -o full_export.sql
```

### 3. Create PostgreSQL Database
```sql
CREATE DATABASE pocketbase;
```

### 4. Import to PostgreSQL
```bash
psql -U user -d pocketbase -f schema.sql
psql -U user -d pocketbase -f full_export.sql  # if including data
```

### 5. Start PocketBase with PostgreSQL
```bash
export PB_DB_TYPE=postgres
export PB_DB_URL="postgres://user:password@localhost:5432/pocketbase"
./pocketbase serve
```

## Known Limitations & Caveats ⚠️

### 1. File Storage
- File storage still uses the local filesystem (`pb_data/storage/`)
- Consider using S3-compatible storage for PostgreSQL deployments
- File paths are stored in the database, files themselves are not

### 2. Backups
- The built-in backup system is designed for SQLite
- For PostgreSQL, use `pg_dump` and standard PostgreSQL backup tools

### 3. SQL Dialect Translation
- Simple string replacement for SQL translation has limitations:
  - May incorrectly match patterns inside string literals or comments
  - Custom-formatted migration SQL might not translate correctly
  - For complex migrations, consider writing database-specific SQL files

### 4. Testing Coverage
- Core functionality is tested
- Integration tests with a live PostgreSQL instance are recommended
- Some edge cases in SQL translation may need refinement

### 5. VACUUM Operations
- `app.Vacuum()` works on both databases
- PostgreSQL VACUUM behavior differs from SQLite's

### 6. Transaction Isolation
- SQLite uses serializable isolation by default
- PostgreSQL uses read committed by default
- Behavior may differ in high-concurrency scenarios

## Next Steps for Production Readiness

### Short Term (Recommended Before Production Use)
1. **Test with your specific use case** - Run your application against PostgreSQL
2. **Benchmark performance** - Compare SQLite vs PostgreSQL for your workload
3. **Test migrations** - Verify schema/data export and import
4. **Test concurrent access** - PostgreSQL excels here but test your patterns

### Medium Term (Community Contributions Welcome)
1. **Expand test coverage** - Add PostgreSQL-specific integration tests
2. **Migration tool improvements** - Better handling of edge cases in SQL translation
3. **Connection pool tuning** - Document optimal settings for different workloads
4. **Performance benchmarks** - Comprehensive benchmarks for common operations

### Long Term
1. **Full feature parity verification** - Systematic testing of all PocketBase features
2. **Documentation expansion** - PostgreSQL-specific deployment guides
3. **Tool integration** - Better integration with PostgreSQL ecosystem tools

## Testing PostgreSQL Support

### Run PostgreSQL-specific tests
```bash
go test ./core/... -run "TestPostgres|TestDBType|TestTranslateSQL|TestDialect" -v
```

### Run full test suite (requires SQLite - default)
```bash
go test ./...
```

### Integration testing with PostgreSQL
```bash
# Start PostgreSQL (Docker example)
docker run -d \
  --name pocketbase-postgres \
  -e POSTGRES_USER=pocketbase \
  -e POSTGRES_PASSWORD=pocketbase \
  -e POSTGRES_DB=pocketbase \
  -p 5432:5432 \
  postgres:16

# Run PocketBase against PostgreSQL
PB_DB_TYPE=postgres \
PB_DB_URL="postgres://pocketbase:pocketbase@localhost:5432/pocketbase?sslmode=disable" \
./pocketbase serve
```

## Architecture References

- `ARCHITECTURE_MULTI_DB.md` - Detailed architecture design document
- `DATABASE_CONFIG.md` - Configuration reference
- `PR_SUMMARY.md` - Implementation summary

## When to Use PostgreSQL vs SQLite

### Choose SQLite when:
- Single server deployment
- Lower traffic (< 100 concurrent users)
- Simplicity is priority
- Development/testing environments
- Embedded applications

### Choose PostgreSQL when:
- High write concurrency requirements
- Horizontal scaling needed (read replicas)
- Existing PostgreSQL infrastructure
- Advanced querying features needed
- Multi-node deployment required

## Reporting Issues

If you encounter issues with PostgreSQL support:

1. Check this STATUS.md for known limitations
2. Try reproducing with SQLite to isolate if it's PostgreSQL-specific
3. Include your PostgreSQL version and connection parameters (without credentials)
4. Provide minimal reproduction steps

---

**Last Updated**: January 2026  
**PocketBase Version**: Development (PostgreSQL support branch)
