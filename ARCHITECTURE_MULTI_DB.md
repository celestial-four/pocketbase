# Multi-Database Backend Support - Architecture Design

## Overview

This document describes the architectural foundation for supporting multiple database backends in PocketBase, starting with SQLite (default) and planned PostgreSQL support.

## Design Principles

### 1. Zero Breaking Changes
- SQLite remains the default and only change to existing behavior
- All existing code continues to work without modification
- New features are strictly opt-in via explicit configuration

### 2. Production-Ready Foundation
- Infrastructure is production-ready and fully tested
- Clear separation between "implemented" and "planned" features
- Graceful degradation with clear error messages

### 3. Configuration Hierarchy
```
CLI Flags (highest priority)
    ↓
Environment Variables
    ↓
Config Struct Defaults
    ↓
System Defaults (SQLite)
```

## Implementation

### Core Components

#### 1. DatabaseConfig (`core/db_config.go`)
```go
type DatabaseConfig struct {
    Type DatabaseType  // "sqlite" or "postgres"
    URL  string        // Connection URL (for postgres)
}
```

**Features:**
- Type-safe database type enumeration
- Environment variable reading (`PB_DB_TYPE`, `PB_DB_URL`)
- Validation with clear error messages
- Helper methods for type checking

#### 2. BaseAppConfig Extension
Added `DBConfig *DatabaseConfig` field to support database configuration while maintaining backward compatibility.

#### 3. CLI Integration (`pocketbase.go`)
New flags:
- `--dbType`: Specify database type
- `--dbURL`: Specify connection URL

### Configuration Resolution

The system resolves database configuration in this order:

1. **CLI Flags**: `--dbType` and `--dbURL`
2. **Environment Variables**: `PB_DB_TYPE` and `PB_DB_URL`
3. **Defaults**: SQLite with standard file paths

```go
// Example resolution logic
dbConfig := &core.DatabaseConfig{}

if pb.dbTypeFlag != "" {
    dbConfig.Type = core.DatabaseType(strings.ToLower(pb.dbTypeFlag))
    dbConfig.URL = pb.dbURLFlag
} else {
    envConfig := core.GetDatabaseConfigFromEnv()
    dbConfig.Type = envConfig.Type
    dbConfig.URL = envConfig.URL
}

// Validate and fall back to SQLite if invalid
if err := dbConfig.Validate(); err != nil {
    slog.Warn("Invalid database configuration, falling back to SQLite", "error", err)
    dbConfig = &core.DatabaseConfig{Type: core.DatabaseTypeSQLite}
}
```

## Testing Strategy

### Test Coverage
1. **Unit Tests** (`core/db_config_test.go`):
   - Environment variable reading
   - Configuration validation
   - Helper method behavior
   - Integration with BaseApp

2. **Integration Tests**:
   - CLI flag parsing
   - Configuration priority
   - Fallback behavior
   - Backward compatibility

3. **Full Test Suite**:
   - All existing tests pass unchanged
   - No regressions in SQLite functionality

### Test Results
```
✅ All core tests pass
✅ All API tests pass
✅ All form tests pass
✅ All integration tests pass
✅ Build succeeds
✅ CLI flags appear in help
✅ Application starts correctly
```

## PostgreSQL Support - Future Work

### Current Status
PostgreSQL support is **not implemented**. The `PostgreSQLDBConnect` function returns an error:

```go
func PostgreSQLDBConnect(dbURL string) (*dbx.DB, error) {
    return nil, fmt.Errorf("PostgreSQL support is not yet implemented - SQLite is the only supported database")
}
```

### What's Needed

#### 1. SQL Dialect Translation
PocketBase uses many SQLite-specific features that need translation:

**Examples:**
```sql
-- SQLite
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL
);

-- PostgreSQL equivalent needed
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT ('r' || encode(gen_random_bytes(7), 'hex')) NOT NULL
);
```

**Other SQLite-specific features:**
- `strftime()` → PostgreSQL date functions
- `AUTOINCREMENT` → `SERIAL` or `BIGSERIAL`
- `JSON` functions differ significantly
- `PRAGMA` statements need alternatives
- Index syntax differences

#### 2. Schema Migrations
Create PostgreSQL-compatible versions of all migrations in `/migrations`:
- Data type conversions
- Constraint translations
- Index definitions
- Trigger replacements (if any)

#### 3. Query Builder Adaptation
The `github.com/pocketbase/dbx` package needs to handle:
- Parameter placeholders (`?` vs `$1, $2, ...`)
- Quote identifiers differently
- Different SQL functions

#### 4. Connection Management
- Connection pooling (already in place, needs tuning)
- Transaction isolation levels
- Prepared statement caching
- Connection health checks

#### 5. Data Type Mapping
| SQLite | PostgreSQL | Notes |
|--------|-----------|-------|
| TEXT | VARCHAR/TEXT | Similar |
| INTEGER | INTEGER/BIGINT | Range differences |
| REAL | DOUBLE PRECISION | Similar |
| BLOB | BYTEA | Binary data |
| JSON | JSONB | PostgreSQL more powerful |
| BOOLEAN | BOOLEAN | SQLite uses 0/1 |

#### 6. Concurrency & Locking
- SQLite uses file-level locking
- PostgreSQL uses row-level locking
- Need to test concurrent writes extensively
- Transaction isolation testing

#### 7. Testing Requirements
- Parallel test suite for both databases
- Performance benchmarks
- Migration testing (SQLite → PostgreSQL)
- Multi-node testing for PostgreSQL
- Failure scenario testing

### Migration Path (Planned)

1. **Phase 1**: Schema export tool
   ```bash
   pocketbase export-schema --format postgres > schema.sql
   ```

2. **Phase 2**: Data migration tool
   ```bash
   pocketbase migrate sqlite-to-postgres \
     --source ./pb_data/data.db \
     --target postgres://user:pass@localhost/dbname
   ```

3. **Phase 3**: Hybrid mode (read from SQLite, write to both)
4. **Phase 4**: Switch to PostgreSQL-only

## Risks & Trade-offs

### Risks
1. **Complexity**: Supporting two databases increases maintenance burden
2. **Testing**: Must test all features on both databases
3. **Performance**: PostgreSQL may have different performance characteristics
4. **File Storage**: PostgreSQL doesn't support embedded storage like SQLite

### Trade-offs
1. **Simplicity vs Scalability**: SQLite is simpler, PostgreSQL scales better
2. **Deployment**: SQLite is zero-config, PostgreSQL needs infrastructure
3. **Development**: SQLite is great for dev, PostgreSQL for production

### Mitigation Strategies
1. Keep SQLite as the recommended default for small/medium deployments
2. Clearly document when PostgreSQL is beneficial
3. Provide migration tools and clear documentation
4. Maintain feature parity between databases
5. Use database abstraction to minimize database-specific code

## Performance Considerations

### SQLite Strengths
- Extremely fast for read-heavy workloads
- Zero network latency
- Simple deployment
- Great for embedded use cases

### PostgreSQL Strengths
- Better write concurrency
- Horizontal scaling with replicas
- Advanced querying capabilities
- Better for multi-user scenarios

### Benchmark Goals (Future)
```
Scenario                SQLite      PostgreSQL    Notes
-------------------------------------------------------------
Single user reads       Very Fast   Fast          Network overhead
Concurrent writes       Moderate    Fast          PostgreSQL wins
Large datasets (>10GB)  Slow        Fast          Memory constraints
Multi-node              N/A         Fast          SQLite can't cluster
```

## Security Considerations

### SQLite
- File permissions critical
- Encryption via SQLCipher (if needed)
- Backup security important

### PostgreSQL
- Network security (SSL/TLS)
- Authentication (passwords, certificates)
- Authorization (row-level security)
- Connection security
- Audit logging

## Deployment Scenarios

### Recommended: SQLite
- Personal projects
- Small teams (<10 users)
- Low-to-moderate traffic
- Single server deployment
- Embedded applications

### Future: PostgreSQL
- Enterprise deployments
- High-concurrency requirements
- Multiple replicas needed
- Advanced querying needed
- Regulatory compliance requirements

## Non-Goals

1. **MySQL/MariaDB Support**: Not planned
2. **MongoDB Support**: Not planned
3. **Multi-database at runtime**: Pick one database type per deployment
4. **Automatic migration**: Requires manual planning and execution
5. **Database-specific features**: Maintain a common feature set

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [dbx Package](https://github.com/pocketbase/dbx)
- [Go database/sql](https://golang.org/pkg/database/sql/)

## Contributing

Before implementing PostgreSQL support:

1. **Open an issue** to discuss the implementation approach
2. **Review this architecture document** and propose changes if needed
3. **Create a proof-of-concept** for the most challenging parts (SQL translation)
4. **Ensure backward compatibility** throughout
5. **Add comprehensive tests** for both databases

## Conclusion

This implementation provides a **solid, production-ready foundation** for multi-database support while maintaining PocketBase's core value proposition: simplicity and ease of use.

The infrastructure is in place, fully tested, and ready for PostgreSQL implementation when the time is right. The design ensures that adding PostgreSQL support will not break existing SQLite-based deployments and that users can choose the right database for their use case.
