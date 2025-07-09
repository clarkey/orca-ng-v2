# GORM Migration Guide

This document describes the changes made to migrate ORCA from direct PostgreSQL (pgx) to GORM for database abstraction.

## Overview

The migration to GORM enables:
- **Multi-database support**: PostgreSQL, MySQL, SQLite, SQL Server
- **Automatic migrations**: Schema management via GORM AutoMigrate
- **Simplified queries**: Object-relational mapping
- **Database portability**: No PostgreSQL-specific features

## Key Changes

### 1. Database Models
All models have been recreated with GORM tags in `/backend/internal/models/gorm/`:
- `user.go` - User and Session models
- `certificate_authority.go` - Certificate Authority model
- `cyberark_instance.go` - CyberArk Instance model  
- `operation.go` - Operation model
- `pipeline_config.go` - Pipeline Configuration model

### 2. Handlers
New GORM-based handlers in `/backend/internal/handlers/`:
- `auth_gorm.go` - Authentication handler
- `certificate_authorities_gorm.go` - Certificate management
- `cyberark_instances_gorm.go` - CyberArk instance management
- `operations_gorm.go` - Operations handler (pipeline metrics removed)

### 3. Services
- `certificate_manager_gorm.go` - Certificate pool management using GORM

### 4. Middleware
- `auth_gorm.go` - Authentication middleware using GORM

### 5. Main Application
- `main_gorm.go` - Updated server initialization with GORM

## Migration Steps

### 1. Update Dependencies
```bash
# Replace go.mod with go_gorm.mod
mv backend/go.mod backend/go.mod.backup
mv backend/go_gorm.mod backend/go.mod

# Download new dependencies
cd backend
go mod download
```

### 2. Environment Variables
Set the database driver (defaults to PostgreSQL if not set):
```bash
export DATABASE_DRIVER=postgres  # Options: postgres, mysql, sqlite, sqlserver
```

### 3. Database Migration
GORM will automatically create/update tables on startup using AutoMigrate.

For existing databases:
1. **Backup your database first!**
2. The GORM models match the existing schema
3. AutoMigrate will add any missing columns/indexes

### 4. Update Application Entry Point
```bash
# Use the new GORM-based main file
mv backend/cmd/orca/main.go backend/cmd/orca/main_pgx.go
mv backend/cmd/orca/main_gorm.go backend/cmd/orca/main.go
```

## Database-Specific Configuration

### PostgreSQL
```bash
export DATABASE_DRIVER=postgres
export DATABASE_URL="postgres://user:password@localhost:5432/orca?sslmode=disable"
```

### MySQL
```bash
export DATABASE_DRIVER=mysql
export DATABASE_URL="user:password@tcp(localhost:3306)/orca?parseTime=true"
```

### SQLite
```bash
export DATABASE_DRIVER=sqlite
export DATABASE_URL="orca.db"  # File path for SQLite database
```

### SQL Server
```bash
export DATABASE_DRIVER=sqlserver
export DATABASE_URL="sqlserver://user:password@localhost:1433?database=orca"
```

## Breaking Changes

1. **Removed Features**:
   - Pipeline metrics functionality (not required yet)
   - PostgreSQL-specific features (tsvector search, JSONB)
   - Direct SQL queries

2. **ID Generation**:
   - ULIDs are now generated in application code (BeforeCreate hooks)
   - No database-level ID generation

3. **JSON Fields**:
   - Changed from JSONB to JSON for cross-database compatibility
   - Using `gorm.io/datatypes` for JSON handling

4. **Cascade Deletes**:
   - Handled by GORM associations instead of database constraints

## Testing

1. **Unit Tests**: Use SQLite in-memory database
2. **Integration Tests**: Test with your target database
3. **Migration Testing**: Always test on a copy of production data

## Rollback Plan

If you need to rollback:
1. Restore the original `main.go` and `go.mod`
2. Restore database from backup if schema was modified
3. Restart the application

## Performance Considerations

- GORM adds minimal overhead for most operations
- Complex queries may need optimization
- Use preloading to avoid N+1 queries
- Monitor query performance in production

## Future Enhancements

1. Add database-specific optimizations via GORM hooks
2. Implement soft deletes where appropriate
3. Add query result caching
4. Enhanced migration tooling