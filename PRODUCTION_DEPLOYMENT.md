# ORCA Production Deployment Guide

## Security Requirements

### 1. Generate Required Secret Keys

Before deploying ORCA in production, you MUST generate secure random keys:

```bash
# Generate ENCRYPTION_KEY (32 bytes, base64 encoded)
openssl rand -base64 32

# Generate SESSION_SECRET (32 bytes, base64 encoded)
openssl rand -base64 32
```

**IMPORTANT**: 
- Store these keys securely (e.g., in a secrets management system)
- Never commit these keys to version control
- Use different keys for each environment
- Rotate keys periodically

### 2. Environment Variables

Required environment variables for production:

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/database?sslmode=require
DATABASE_DRIVER=postgres  # Options: postgres, mysql, sqlite, sqlserver

# Security Keys (MUST be generated securely)
ENCRYPTION_KEY=<your-32-byte-key>  # For encrypting CyberArk passwords
SESSION_SECRET=<your-32-byte-key>  # For session encryption

# Application Environment
APP_ENV=production
LOG_LEVEL=info

# Optional
GORM_DEBUG=false  # Set to true only for debugging
```

## Initial Admin Setup

### Option 1: Environment Variable Configuration (Recommended)

Set initial admin credentials via environment variables:

```bash
# Add these to your deployment
INITIAL_ADMIN_USERNAME=admin
INITIAL_ADMIN_PASSWORD=<secure-generated-password>
```

### Option 2: First-Run CLI Setup

Use the CLI to create the first admin user:

```bash
# After deployment, run:
orca-cli admin create --username admin --email admin@company.com
# The CLI will prompt for a secure password
```

### Option 3: Kubernetes Secrets

For Kubernetes deployments:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: orca-secrets
type: Opaque
data:
  encryption-key: <base64-encoded-key>
  session-secret: <base64-encoded-key>
  initial-admin-password: <base64-encoded-password>
```

## Docker Compose Production Example

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - orca_network
    restart: unless-stopped

  backend:
    image: orca-backend:latest
    environment:
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=require
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      SESSION_SECRET: ${SESSION_SECRET}
      APP_ENV: production
      LOG_LEVEL: info
      INITIAL_ADMIN_USERNAME: ${INITIAL_ADMIN_USERNAME}
      INITIAL_ADMIN_PASSWORD: ${INITIAL_ADMIN_PASSWORD}
    depends_on:
      - postgres
    networks:
      - orca_network
    restart: unless-stopped

  frontend:
    image: orca-frontend:latest
    environment:
      BACKEND_URL: http://backend:8080
    depends_on:
      - backend
    networks:
      - orca_network
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - frontend
      - backend
    networks:
      - orca_network
    restart: unless-stopped

volumes:
  postgres_data:

networks:
  orca_network:
    driver: bridge
```

## Security Checklist

- [ ] Generate unique ENCRYPTION_KEY using `openssl rand -base64 32`
- [ ] Generate unique SESSION_SECRET using `openssl rand -base64 32`
- [ ] Set strong initial admin password
- [ ] Enable SSL/TLS for database connections
- [ ] Configure HTTPS with valid certificates
- [ ] Set up firewall rules
- [ ] Enable audit logging
- [ ] Configure backup strategy
- [ ] Set up monitoring and alerting
- [ ] Review and harden PostgreSQL configuration
- [ ] Implement rate limiting
- [ ] Configure CORS properly for your domain
- [ ] Set up log rotation
- [ ] Document emergency access procedures

## Post-Deployment Steps

1. **Change default admin password immediately**
   ```bash
   # Login as admin
   curl -X POST https://your-domain/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "<initial-password>"}'
   
   # Use the web UI to change password
   ```

2. **Create additional admin users**
   - Avoid using the default 'admin' account for daily operations
   - Create named accounts for each administrator

3. **Configure CyberArk integration**
   - Add CyberArk instances
   - Upload necessary CA certificates
   - Test connections

4. **Set up monitoring**
   - Health endpoint: `GET /health`
   - Metrics endpoint: Configure Prometheus/Grafana

## Backup Strategy

1. **Database Backups**
   ```bash
   # PostgreSQL backup
   pg_dump -h localhost -U orca -d orca_db > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Encryption Key Backup**
   - Store encryption keys in a secure key management system
   - Keep offline backups in a secure location
   - Document key rotation procedures

## Troubleshooting

### Forgot Admin Password

If you lose access to the admin account:

1. Access the database directly
2. Create a new admin user:
   ```sql
   INSERT INTO users (id, username, password_hash, is_active, is_admin, created_at, updated_at)
   VALUES ('usr_' || gen_random_uuid(), 'recovery-admin', '<bcrypt-hash>', true, true, NOW(), NOW());
   ```

3. Or use the CLI recovery command:
   ```bash
   orca-cli admin reset --username admin
   ```

### Key Rotation

To rotate encryption keys:

1. Decrypt all encrypted data with old key
2. Re-encrypt with new key
3. Update ENCRYPTION_KEY environment variable
4. Restart application

## Compliance Considerations

- Log all authentication attempts
- Implement session timeout policies
- Enable audit trails for all CyberArk operations
- Configure password complexity requirements
- Implement account lockout policies
- Regular security assessments

## Support

For production support:
- Check logs: `docker logs orca-backend`
- Health check: `curl https://your-domain/health`
- Database status: Check PostgreSQL logs
- Report issues: [GitHub Issues](https://github.com/your-org/orca/issues)