# ORCA - Orchestration for CyberArk

## Project Overview
ORCA is an enterprise application that provides orchestration and management capabilities for CyberArk's self-hosted Privileged Access Management (PAM) system. It abstracts and aggregates permissions on safes within the CyberArk Vault into logical 'safe access roles', providing a modern interface for CyberArk administrators.

### Key Features
- **Multi-Instance Management**: Support for multiple CyberArk PVWA instances
- **Operations Pipeline**: Asynchronous task processing with priority queues
- **Safe Management**: Provision, modify, and delete safes
- **Access Control**: Grant and revoke permissions with role-based abstractions
- **Data Synchronization**: Periodic refresh of safes, users, and groups from CyberArk
- **Real-time Monitoring**: Pipeline metrics and operation status tracking
- **IGA Integration**: Designed for integration with SailPoint IIQ/IdentityNow
- **RESTful API**: All CyberArk communication via REST API

### Technical Stack
- **Backend**: Go 1.23 with Gin framework (API-first approach)
- **Database**: PostgreSQL 16 with pgx/v5 driver
- **Frontend**: React 18.3 SPA embedded in Go binary
  - Vite 6 build tool
  - TypeScript
  - Tailwind CSS v4
  - Shadcn UI (Radix UI based components)
  - TanStack Query (React Query) for state management
  - React Router v7
  - React Hook Form + Zod validation
- **Authentication**: Session-based with Argon2id password hashing
- **Encryption**: AES-256-GCM for sensitive data
- **IDs**: ULID with semantic prefixes (usr_, ses_, cai_, op_, cfg_)
- **CLI**: Cobra-based admin tool with cross-platform session storage
- **Logging**: Logrus with structured logging
- **Configuration**: Viper for config management

### Architecture Notes
- **Target Users**: CyberArk Administrators (not end users)
- **Design Philosophy**: Improves upon vendor SCIM solutions
- **Deployment**: Single binary with embedded frontend for production
- **Async Processing**: Pipeline architecture for long-running operations
- **Multi-tenancy**: Support for multiple CyberArk environments
- **Session Management**: Concurrent session support for CyberArk

### Build Tags
The project uses Go build tags to control compilation:
- **`-tags dev`**: Development mode - skips embedding frontend files for faster builds
- **No tags**: Production mode - embeds the frontend `dist` folder into the binary

This is why development commands use `go run -tags dev` while production uses the pre-built binary.

### Current Implementation Status

#### ✅ Completed
- Authentication system (web + CLI)
- Session management with secure storage
- CyberArk instance management (CRUD + connection testing)
- Operations pipeline framework
- Collapsible sidebar with persistent state
- Real-time pipeline metrics
- Base UI structure with routing

#### 🚧 In Progress
- Dashboard implementation
- Safe management operations
- User and group synchronization
- Access role definitions
- Settings pages (SSO, notifications, etc.)

#### 📋 Planned
- IGA integration endpoints
- Audit logging system
- Advanced search and filtering
- Bulk operations support
- Role-based access control
- API rate limiting

### Development Setup

⚠️ **IMPORTANT: All services run in Docker containers!** ⚠️
- NEVER run `npm run dev` or `go run` directly on the host
- ALWAYS use `docker-compose` commands
- Frontend and backend are containerized services

```bash
# Start all services (frontend, backend, postgres)
docker-compose up

# Backend runs on http://localhost:8080 (in container)
# Frontend dev server runs on http://localhost:5173 (in container, NOT 5175)
# PostgreSQL on localhost:5432 (in container)

# View logs
docker-compose logs -f frontend  # Frontend logs
docker-compose logs -f backend   # Backend logs

# Restart a specific service
docker-compose restart frontend  # To pick up frontend changes
docker-compose restart backend   # To pick up backend changes

# Default admin credentials
Username: admin
Password: admin123 (CHANGE THIS!)
```

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string
- `SESSION_SECRET`: Session encryption key (32 bytes)
- `ENCRYPTION_KEY`: Data encryption key (32 bytes)
- `LOG_LEVEL`: debug, info, warn, error
- `APP_ENV`: development, production

### Security Considerations
- Change default admin password immediately
- Generate strong SESSION_SECRET and ENCRYPTION_KEY
- Configure CORS for production
- Implement HTTPS in production
- Review and implement rate limiting
- Enable audit logging for compliance

### Admin Password Management

The default admin credentials are:
- Username: `admin`
- Password: `admin123`

**⚠️ IMPORTANT: Change this immediately after first deployment!**

To reset the admin password:

1. **Using the built-in password reset flag (recommended)**:
   ```bash
   # Development: uses -tags dev to skip frontend embedding
   docker-compose exec backend sh -c "go run -tags dev ./cmd/orca --reset-password='your-new-password'"
   
   # Production: use the pre-built binary
   docker-compose exec backend ./orca-server --reset-password='your-new-password'
   ```

2. **Using the CLI with local database access**:
   ```bash
   # Build the CLI first
   docker-compose exec backend go build -o orca-cli cmd/orca-cli/main.go
   
   # Reset password with DATABASE_URL
   docker-compose exec backend sh -c "DATABASE_URL='postgres://orca:orca@postgres:5432/orca?sslmode=disable' ./orca-cli user reset-password --username admin --local"
   ```

3. **Generate secure passwords**:
   ```bash
   # Build orca-keygen
   docker-compose exec backend go build -o orca-keygen cmd/orca-keygen/main.go
   
   # Generate a secure password
   docker-compose exec backend ./orca-keygen password
   ```

Note: `orca-keygen` generates secure random passwords and keys, not password hashes. It's useful for:
- Generating secure admin passwords
- Creating SESSION_SECRET and ENCRYPTION_KEY values
- Producing all required security keys for production deployment

### Testing Strategy
- Unit tests: `make test`
- Integration tests: `make test-integration`
- E2E tests: (planned)
- Security scanning: (planned)

### Important Notes for Development
1. All database IDs use ULID with semantic prefixes
2. Frontend is embedded in binary for production builds
3. Operations are async - use the pipeline for long-running tasks
4. Always validate input on both frontend (Zod) and backend
5. Follow existing patterns for new features
6. Use structured logging with appropriate log levels
7. Maintain backwards compatibility for API changes

### Common Development Commands

⚠️ **REMINDER: Everything runs in Docker!** ⚠️

```bash
# Docker commands (use these!)
docker-compose up                  # Start all services
docker-compose down                # Stop all services
docker-compose restart frontend    # Restart frontend after changes
docker-compose restart backend     # Restart backend after changes
docker-compose exec backend bash   # Shell into backend container
docker-compose exec frontend sh    # Shell into frontend container

# If you need to run commands, do it INSIDE the container:
docker-compose exec backend go run cmd/cli/main.go
docker-compose exec frontend npm run lint

# Build commands (still use docker-compose)
docker-compose build frontend      # Rebuild frontend image
docker-compose build backend       # Rebuild backend image

# CLI Usage
orca-cli login                     # Authenticate
orca-cli status                    # Check connection
orca-cli users list               # List users
orca-cli config get               # View config

# Admin Password Reset (built into main binary)
# Note: -tags dev skips frontend embedding for faster execution in development
docker-compose exec backend sh -c "go run -tags dev ./cmd/orca --reset-password='new-password'"
docker-compose exec backend sh -c "go run -tags dev ./cmd/orca --reset-password='new-password' --username='john.doe'"

# Generate secure passwords with orca-keygen
docker-compose exec backend go build -o orca-keygen cmd/orca-keygen/main.go
docker-compose exec backend ./orca-keygen password     # 16-char password
docker-compose exec backend ./orca-keygen password 24  # 24-char password
docker-compose exec backend ./orca-keygen all          # All production keys
```

### Code Style Guidelines
- Go: Follow standard Go conventions
- TypeScript: Use TypeScript strict mode
- React: Functional components with hooks
- CSS: Tailwind utility classes, avoid custom CSS
- Git: Conventional commits (feat:, fix:, chore:, etc.)
- Comments: Explain "why" not "what"