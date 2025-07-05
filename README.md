# ORCA - Orchestration for CyberArk

ORCA is an application that wraps around the CyberArk self-hosted Privileged Access Management (PAM) system. It abstracts and aggregates permissions on safes within the CyberArk Vault into logical 'safe access roles'.

## Features

- Multi-instance CyberArk management
- Safe provisioning and access management
- Integration with IGA products (SailPoint IIQ/IdentityNow)
- Session-based authentication with Argon2id
- Web UI and CLI administration tools

## Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with TypeScript, Vite, Shadcn UI, and TailWind CSS v4
- **Database**: PostgreSQL
- **Authentication**: Session-based with secure cookie storage

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)
- Node.js 20+ (for local development)

### Running with Docker Compose

1. Clone the repository
2. Start the services:
   ```bash
   make build
   make run
   ```

3. Access the application at http://localhost:5173

Default credentials:
- Username: `admin`
- Password: `admin`

**Important**: Change the default password immediately after first login.

### Building from Source

```bash
# Build both server and CLI
./scripts/build.sh

# Or build individually
cd backend
go build -o orca ./cmd/orca
go build -o orca-cli ./cmd/orca-cli
```

## CLI Usage

The ORCA CLI provides administrative access to the system:

```bash
# Login to ORCA
./orca-cli login -s http://localhost:8080

# Check session status
./orca-cli status

# Logout
./orca-cli logout
```

Session tokens are stored securely in:
- macOS: `~/Library/Application Support/orca-cli/`
- Linux: `~/.config/orca-cli/`
- Windows: `%APPDATA%\orca-cli\`

## Development

### Backend Development

The backend uses Air for hot reloading:

```bash
cd backend
air
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Database Migrations

Migrations are automatically applied when the PostgreSQL container starts. Additional migrations can be added to the `migrations/` directory.

## Project Structure

```
├── backend/
│   ├── cmd/
│   │   ├── orca/          # Main server
│   │   └── orca-cli/      # CLI tool
│   ├── internal/          # Internal packages
│   └── pkg/               # Reusable packages
├── frontend/              # React application
├── migrations/            # Database migrations
└── docker-compose.yml     # Development environment
```

## Security Notes

- All IDs use ULID with appropriate prefixes (e.g., `usr_`, `ses_`)
- Passwords are hashed using Argon2id
- Sessions expire after 24 hours by default
- HTTPS should be used in production

## License

[License information to be added]