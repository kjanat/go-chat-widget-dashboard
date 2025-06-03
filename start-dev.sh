#!/bin/bash

echo "Starting Go Chat Widget Dashboard with air..."
cd /workspaces/go-chat-widget-dashboard

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "Installing air..."
    go install github.com/cosmtrek/air@latest
fi

# Run with air
echo "Starting development server on port 3000..."
air -c .air.toml
