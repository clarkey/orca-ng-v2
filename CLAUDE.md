# ORCA - Orchestration for CyberArk

## Project Overview
ORCA is an application that wraps around the CyberArk self-hosted Privileged Access Management (PAM) system. It abstracts/aggregates permissions on safes within the CyberArk Vault into logical 'safe access roles'.

### Key Features
- Provision safes
- Manage safe access
- Multi-instance CyberArk management
- Integration with IGA products (SailPoint IIQ/IdentityNow)
- Periodic data refresh from CyberArk (safes, users, groups)
- All communication via CyberArk REST API

### Technical Stack
- **Backend**: Golang with GIN framework (API-first approach)
- **Logging**: Logrus
- **Frontend**: SPA embedded in Go binary
  - Shadcn UI
  - TailWind v4
  - Vite
  - TypeScript
- **Database**: PostgreSQL
- **Authentication**: Session-based with Argon2id
- **IDs**: ULID with appropriate prefixes (like Stripe)
- **CLI**: Separate admin CLI with cross-platform session token storage

### Architecture Notes
- Not a user-facing application
- Designed for CyberArk Administrators
- Improves upon vendor SCIM solutions
- All components use latest LTS versions

### Development Setup
- Docker Compose for local development
- Session-based authentication backed by PostgreSQL
- CLI supports macOS, Windows, and Linux session storage