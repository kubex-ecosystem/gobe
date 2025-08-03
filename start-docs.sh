#!/bin/bash

# KbxCtl API Documentation Server
# This script starts the comprehensive API documentation server

echo "🚀 Starting KbxCtl API Documentation Server..."
echo ""
echo "This will start a beautiful documentation portal with:"
echo "  📚 Beautiful HTML documentation"
echo "  📊 JSON API documentation"
echo "  🗺️  Complete routes mapping"
echo "  🎯 Live API endpoints"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Navigate to the project directory
cd "$(dirname "$0")/../.." || exit 1

# Build and run the documentation server
echo "🔨 Building documentation server..."
if go build -o gobe-docs cmd/api-docs/main.go; then
    echo "✅ Build successful!"
    echo ""
    echo "🌟 Starting documentation server..."
    echo "📱 Open your browser to:"
    echo "   🏠 Main API: http://localhost:8080/"
    echo "   📚 Beautiful docs: http://localhost:8080/docs"
    echo "   📊 API JSON: http://localhost:8080/api/docs"
    echo "   🗺️  Routes: http://localhost:8080/api/routes"
    echo ""
    echo "Press Ctrl+C to stop the server"
    echo ""
    
    ./gobe-docs
else
    echo "❌ Build failed. Please check for errors."
    exit 1
fi
