#!/bin/bash
cd /workspaces/go-chat-widget-dashboard
export PATH=$PATH:/go/bin
echo "Starting templ generation..."
templ generate 2>&1 | tee templ_output.log
echo "Exit code: $?"
echo "Output saved to templ_output.log"
