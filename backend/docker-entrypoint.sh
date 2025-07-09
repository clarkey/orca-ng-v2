#!/bin/sh
set -e

# Function to generate secure random string
generate_secure_key() {
    openssl rand -base64 32 | tr -d '\n'
}

# Check if running in production
if [ "$APP_ENV" = "production" ]; then
    echo "Running in production mode - validating configuration..."
    
    # Check encryption key
    if [ -z "$ENCRYPTION_KEY" ]; then
        echo "ERROR: ENCRYPTION_KEY is not set!"
        echo "Generate one with: openssl rand -base64 32"
        exit 1
    fi
    
    # Check session secret
    if [ -z "$SESSION_SECRET" ]; then
        echo "ERROR: SESSION_SECRET is not set!"
        echo "Generate one with: openssl rand -base64 32"
        exit 1
    fi
    
    # Warn about default values
    if [ "$ENCRYPTION_KEY" = "development-key-do-not-use-in-prod-32bytes!!" ]; then
        echo "ERROR: Using development ENCRYPTION_KEY in production!"
        exit 1
    fi
    
    if [ "$SESSION_SECRET" = "development-secret-key-32-bytes!!" ]; then
        echo "ERROR: Using development SESSION_SECRET in production!"
        exit 1
    fi
    
    # Check initial admin password
    if [ -z "$INITIAL_ADMIN_PASSWORD" ]; then
        echo "WARNING: INITIAL_ADMIN_PASSWORD not set. A secure password will be generated."
        echo "Check the logs for the generated password - it will only be shown once!"
    fi
else
    echo "Running in development mode"
    
    # Set development defaults if not provided
    if [ -z "$ENCRYPTION_KEY" ]; then
        export ENCRYPTION_KEY="development-key-do-not-use-in-prod-32bytes!!"
    fi
    
    if [ -z "$SESSION_SECRET" ]; then
        export SESSION_SECRET="development-secret-key-32-bytes!!"
    fi
fi

# Wait for database to be ready
if [ -n "$DATABASE_URL" ]; then
    echo "Waiting for database to be ready..."
    
    # Extract host and port from DATABASE_URL
    # This is a simplified version - in production use a proper URL parser
    DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
    DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
    
    if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
        while ! nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; do
            echo "Database not ready, waiting..."
            sleep 2
        done
        echo "Database is ready!"
    fi
fi

# Execute the main application
exec "$@"