#!/bin/bash

# Chat App Startup Script
echo "🚀 Starting Chat Application..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 &> /dev/null; then
    echo "⚠️  PostgreSQL is not running. Please start PostgreSQL service."
    echo "   Ubuntu/Debian: sudo systemctl start postgresql"
    echo "   macOS: brew services start postgresql"
    echo "   Or start manually: pg_ctl -D /usr/local/var/postgres start"
fi

# Check if Redis is running
if ! redis-cli ping &> /dev/null; then
    echo "⚠️  Redis is not running. Please start Redis service."
    echo "   Ubuntu/Debian: sudo systemctl start redis-server"
    echo "   macOS: brew services start redis"
fi

# Set default environment variables if not set
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-123456}
export DB_NAME=${DB_NAME:-chatapp}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export REDIS_HOST=${REDIS_HOST:-localhost}
export REDIS_PORT=${REDIS_PORT:-6379}
export PORT=${PORT:-8080}

echo "📋 Configuration:"
echo "   Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "   Redis: $REDIS_HOST:$REDIS_PORT"
echo "   Port: $PORT"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Run the application
echo "🎯 Starting server on port $PORT..."
echo "🌐 Open http://localhost:$PORT in your browser"
echo ""

go run main.go
