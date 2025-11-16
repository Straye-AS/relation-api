#!/bin/bash

# Development environment setup script

set -e

echo "🚀 Setting up Straye Relation API development environment..."

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker is not installed. Some features may not work."
fi

echo "✅ Prerequisites check passed"

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Install development tools
echo "🔧 Installing development tools..."
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p storage
mkdir -p logs

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please update .env with your configuration"
fi

# Start database with Docker if available
if command -v docker &> /dev/null; then
    echo "🐳 Starting PostgreSQL with Docker Compose..."
    docker-compose up -d postgres
    
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 5
    
    echo "🔄 Running database migrations..."
    go run ./cmd/migrate up
    
    echo "🌱 Seeding database (optional)..."
    if [ -f testdata/seed.sql ]; then
        docker-compose exec -T postgres psql -U relation_user -d relation_db < testdata/seed.sql || true
    fi
fi

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "Next steps:"
echo "  1. Update .env with your configuration (especially Azure AD settings)"
echo "  2. Run 'make run' to start the API server"
echo "  3. Visit http://localhost:8080/swagger for API documentation"
echo "  4. Run 'make test' to run tests"
echo ""
echo "📚 See README.md for more information"

