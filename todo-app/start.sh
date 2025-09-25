#!/bin/bash

# Todo App Startup Script
# This script starts both the backend server and frontend development server

echo "🚀 Starting Todo List Application..."

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a port is in use
port_in_use() {
    lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
}

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command_exists go; then
    echo "❌ Go is not installed. Please install Go 1.18+ first."
    exit 1
fi

if ! command_exists node; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

if ! command_exists npm; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Check if ports are available
if port_in_use 8081; then
    echo "⚠️  Port 8081 is already in use. Please stop the process using this port."
    exit 1
fi

if port_in_use 3000; then
    echo "⚠️  Port 3000 is already in use. Please stop the process using this port."
    exit 1
fi

# Function to cleanup processes on exit
cleanup() {
    echo "🛑 Shutting down servers..."
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
    fi
    echo "👋 Goodbye!"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Start backend server
echo "🔧 Starting backend server..."
cd backend

# Check if go.mod exists
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found in backend directory"
    exit 1
fi

# Download dependencies if needed
go mod download

# Start backend in background
go run main.go &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Check if backend is running
if ! port_in_use 8081; then
    echo "❌ Backend server failed to start"
    exit 1
fi

echo "✅ Backend server started on http://localhost:8081"

# Start frontend server
echo "🔧 Starting frontend development server..."
cd ../frontend

# Check if package.json exists
if [ ! -f "package.json" ]; then
    echo "❌ package.json not found in frontend directory"
    cleanup
    exit 1
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    npm install
fi

# Start frontend in background
npm run dev &
FRONTEND_PID=$!

# Wait a moment for frontend to start
sleep 5

# Check if frontend is running
if ! port_in_use 3000; then
    echo "❌ Frontend server failed to start"
    cleanup
    exit 1
fi

echo "✅ Frontend server started on http://localhost:3000"

echo ""
echo "🎉 Todo List Application is running!"
echo ""
echo "🔗 Frontend: http://localhost:3000"
echo "🔗 Backend:  http://localhost:8081"
echo "🔗 API:      http://localhost:8081/api"
echo "🔗 Health:   http://localhost:8081/health"
echo ""
echo "Press Ctrl+C to stop the servers"

# Wait for processes to finish
wait