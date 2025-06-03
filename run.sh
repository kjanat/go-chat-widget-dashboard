#!/bin/bash
cd /workspaces/go-chat-widget-dashboard
echo "Building application..."
go build -o tmp/main main.go
if [ $? -eq 0 ]; then
    echo "Build successful, starting server on port 3000..."
    ./tmp/main
else
    echo "Build failed"
    exit 1
fi
