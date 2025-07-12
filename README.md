# ORCA - Orchestration for CyberArk

ORCA is an enterprise orchestration platform for CyberArk's Privileged Access Management (PAM) system. It provides a modern interface for managing multiple CyberArk instances, abstracting safe permissions into logical access roles, and streamlining administrative operations.

## 🚀 Features

### Core Capabilities
- **Multi-Instance Management**: Manage multiple CyberArk PVWA instances from a single interface
- **Asynchronous Operations**: Queue-based processing for long-running tasks with real-time status updates
- **Safe Management**: Provision, modify, and delete safes with role-based access abstractions
- **Access Control**: Grant and revoke permissions using logical groupings
- **Data Synchronization**: Automated sync of safes, users, and groups from CyberArk
- **IGA Integration**: Built for integration with SailPoint IIQ/IdentityNow

### User Interface
- **Modern Web UI**: React-based SPA with responsive design
- **Collapsible Sidebar**: Persistent navigation state for better screen utilization
- **Real-time Updates**: Live operation status and pipeline metrics
- **Dark Mode Ready**: Tailwind CSS v4 with theme support
- **CLI Tool**: Command-line interface for automation and scripting

## 🛠️ Technology Stack

- **Backend**: Go 1.23 with Gin web framework
- **Frontend**: React 18.3, TypeScript, Vite 6, Tailwind CSS v4
- **Database**: PostgreSQL 16
- **Authentication**: Session-based with Argon2id hashing
- **UI Components**: Shadcn UI (Radix UI based)
- **State Management**: TanStack Query (React Query)
- **Validation**: Zod schemas with React Hook Form

## 📋 Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.23+ (for local development)
- Node.js 20+ and npm 10+ (for frontend development)
- PostgreSQL 16+ (for production deployment)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/your-org/orca-ng-v2.git
   cd orca-ng-v2
   ```

2. Create environment file:
   ```bash
   cp .env.example .env
   # Edit .env and set secure values for:
   # - SESSION_SECRET (32-byte key)
   # - ENCRYPTION_KEY (32-byte key)
   ```

3. Start the services:
   ```bash
   docker-compose up
   ```

4. Access the application:
   - Web UI: http://localhost:5173
   - API: http://localhost:8080

5. Login with default credentials:
   - Username: `admin`
   - Password: `admin123`

   **⚠️ IMPORTANT**: Change the default password immediately after first login!

### Manual Setup

See [Development Setup](#development-setup) for detailed instructions.

## 💻 CLI Usage

The ORCA CLI provides administrative and automation capabilities:

```bash
# Install the CLI
go install ./backend/cmd/cli

# Login to ORCA
orca-cli login --server http://localhost:8080

# Check connection status
orca-cli status

# List users
orca-cli users list

# Create a new user
orca-cli users create --username john.doe --email john@example.com

# View configuration
orca-cli config get

# Logout
orca-cli logout
```

Session tokens are stored securely:
- **macOS**: `~/Library/Application Support/orca-cli/`
- **Linux**: `~/.config/orca-cli/`
- **Windows**: `%APPDATA%\orca-cli\`

### Resetting the Admin Password

The ORCA server binary includes built-in password reset functionality:

```bash
# In development (using -tags dev to skip frontend embedding for faster execution)
docker-compose exec backend sh -c "go run -tags dev ./cmd/orca --reset-password='new-password'"

# Reset a specific user's password
docker-compose exec backend sh -c "go run -tags dev ./cmd/orca --reset-password='new-password' --username='john.doe'"

# In production (using pre-built binary)
docker-compose exec backend ./orca-server --reset-password='new-password'
```

**Note**: The `-tags dev` flag is used in development to skip embedding frontend files into the binary, making the password reset operation faster. In production, the binary is already built without this flag.

Alternative methods:

```bash
# Using the CLI with local database access
docker-compose exec backend sh -c "go build -o orca-cli cmd/orca-cli/main.go"
docker-compose exec backend sh -c "DATABASE_URL='postgres://orca:orca@postgres:5432/orca?sslmode=disable' ./orca-cli user reset-password --username admin --local"
```

### Generating Secure Passwords

The `orca-keygen` utility can generate secure random passwords:

```bash
# Build the keygen tool
docker-compose exec backend go build -o orca-keygen cmd/orca-keygen/main.go

# Generate a secure password (default 16 characters)
docker-compose exec backend ./orca-keygen password

# Generate a longer password
docker-compose exec backend ./orca-keygen password 24

# Generate all security keys for production
docker-compose exec backend ./orca-keygen all
```

## 🏗️ Project Structure

```
orca-ng-v2/
├── backend/
│   ├── cmd/
│   │   ├── server/        # Main API server
│   │   └── cli/           # CLI application
│   ├── internal/
│   │   ├── api/           # HTTP handlers
│   │   ├── config/        # Configuration
│   │   ├── database/      # Database layer
│   │   ├── models/        # Data models
│   │   ├── pipeline/      # Async processing
│   │   └── services/      # Business logic
│   └── pkg/               # Reusable packages
├── frontend/
│   ├── src/
│   │   ├── api/           # API client
│   │   ├── components/    # React components
│   │   ├── hooks/         # Custom hooks
│   │   ├── pages/         # Route pages
│   │   └── lib/           # Utilities
│   └── public/            # Static assets
├── migrations/            # Database migrations
├── scripts/               # Build and deployment scripts
└── docker-compose.yml     # Development environment
```

## 🔧 Development Setup

### Backend Development

1. Install dependencies:
   ```bash
   cd backend
   go mod download
   ```

2. Set up environment:
   ```bash
   export DATABASE_URL="postgres://orca:orca@localhost:5432/orca?sslmode=disable"
   export SESSION_SECRET="your-32-byte-session-secret-here"
   export ENCRYPTION_KEY="your-32-byte-encryption-key-here"
   ```

3. Run migrations:
   ```bash
   make migrate-up
   ```

4. Start the server with hot reload:
   ```bash
   air
   ```

### Frontend Development

1. Install dependencies:
   ```bash
   cd frontend
   npm install
   ```

2. Start development server:
   ```bash
   npm run dev
   ```

3. Run tests:
   ```bash
   npm run test
   npm run lint
   npm run type-check
   ```

### Building for Production

```bash
# Build everything
./scripts/build.sh

# Or build individually
cd backend
make build

cd ../frontend
npm run build
```

## 🔒 Security Considerations

- **Authentication**: Session-based with secure HTTP-only cookies
- **Password Security**: Argon2id hashing with salt
- **Data Encryption**: AES-256-GCM for sensitive data (CyberArk passwords)
- **ID Generation**: ULID with semantic prefixes for traceability
- **Input Validation**: Comprehensive validation on both frontend and backend
- **CORS**: Configured for development, requires production configuration
- **HTTPS**: Required for production deployment

## 📊 Monitoring and Operations

### Pipeline Metrics
- View real-time metrics at `/pipeline`
- Monitor operation queues and processing rates
- Track success/failure rates by operation type

### Logging
- Structured logging with Logrus
- Log levels: DEBUG, INFO, WARN, ERROR
- Correlation IDs for request tracing

### Health Checks
- `/api/health` - Basic health check
- `/api/health/ready` - Readiness probe (checks DB connection)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using conventional commits (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style
- **Go**: Follow standard Go conventions and `gofmt`
- **TypeScript**: Use strict mode and follow ESLint rules
- **React**: Functional components with TypeScript
- **Git**: Conventional commits (feat:, fix:, docs:, etc.)

## 📝 API Documentation

API documentation is available at `/api/docs` when running in development mode.

Key endpoints:
- `POST /api/auth/login` - Web authentication
- `GET /api/cyberark/instances` - List CyberArk instances
- `GET /api/operations` - List operations with filtering
- `GET /api/pipeline/metrics` - Real-time pipeline metrics

## 🐛 Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 5173, 8080, and 5432 are available
2. **Database connection**: Check DATABASE_URL and PostgreSQL status
3. **Session issues**: Clear cookies and browser cache
4. **Build failures**: Ensure Go 1.23+ and Node.js 20+ are installed

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
```

## 📜 License

[License information to be added]

## 🙏 Acknowledgments

- Built with ❤️ for CyberArk administrators
- Inspired by modern DevOps practices
- Powered by open-source technologies

---

For more detailed information, see [CLAUDE.md](./CLAUDE.md) for development guidelines and architecture details.