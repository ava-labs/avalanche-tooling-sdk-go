#!/bin/bash

# Avalanche Tooling SDK - Quick Start Script
# This script helps you quickly test the server API

set -e

echo "🚀 Avalanche Tooling SDK - Quick Start"
echo "======================================"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Check if session file exists
SESSION_FILE="/Users/raymondsukanto/Desktop/management-session.json"
if [ ! -f "$SESSION_FILE" ]; then
    echo "⚠️  Warning: Session file not found at $SESSION_FILE"
    echo "   The server requires a valid Cubist session file."
    echo "   Please ensure you have the session file in place before running the server."
    echo
fi

# Build the project
echo "📦 Building the project..."
go build ./...
echo "✅ Build completed"
echo

# Function to run server in background
run_server() {
    echo "🖥️  Starting gRPC server on port 8080..."
    go run examples/grpc_server_example.go &
    SERVER_PID=$!
    echo "   Server PID: $SERVER_PID"
    
    # Wait a moment for server to start
    sleep 2
    
    # Check if server is running
    if kill -0 $SERVER_PID 2>/dev/null; then
        echo "✅ Server started successfully"
        return 0
    else
        echo "❌ Failed to start server"
        return 1
    fi
}

# Function to run client
run_client() {
    echo "🔌 Running client example..."
    echo
    go run examples/simple_client_example.go
    echo
    echo "✅ Client example completed"
}

# Function to cleanup
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        echo "🛑 Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo "✅ Server stopped"
    fi
}

# Set up cleanup on exit
trap cleanup EXIT

# Main execution
echo "Choose an option:"
echo "1. Run server only"
echo "2. Run client only (requires server to be running)"
echo "3. Run both server and client"
echo "4. Run complete example (server + detailed client)"
echo
read -p "Enter your choice (1-4): " choice

case $choice in
    1)
        echo "🖥️  Starting server only..."
        run_server
        if [ $? -eq 0 ]; then
            echo "✅ Server is running. Press Ctrl+C to stop."
            wait
        fi
        ;;
    2)
        echo "🔌 Running client only..."
        run_client
        ;;
    3)
        echo "🖥️  Starting server and client..."
        if run_server; then
            run_client
        fi
        ;;
    4)
        echo "🖥️  Starting server and complete example..."
        if run_server; then
            echo "🔌 Running complete client example..."
            go run examples/complete_server_example.go
        fi
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again and choose 1-4."
        exit 1
        ;;
esac

echo
echo "🎉 Quick start completed!"
echo
echo "📚 For more information, see:"
echo "   - SERVER_API_GUIDE.md - Complete API documentation"
echo "   - examples/ - More example code"
echo "   - Makefile - Additional build targets"
