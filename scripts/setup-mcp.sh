#!/bin/bash
# Setup script for FlyFleet MCP configuration
# This script helps you configure the Google API key for Stitch MCP

set -e

echo "🔧 FlyFleet MCP Setup"
echo "===================="
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ Created .env file"
else
    echo "✅ .env file already exists"
fi

echo ""
echo "⚠️  IMPORTANT: You need to set your Google Cloud API Key"
echo ""
echo "To configure your API key, choose one of these options:"
echo ""
echo "Option 1: Add to .env file (project-specific)"
echo "  1. Open .env file in this directory"
echo "  2. Replace 'your-api-key-here' with your actual API key"
echo "  3. Save the file"
echo ""
echo "Option 2: Set as environment variable (system-wide)"
echo "  Add this line to your ~/.bashrc or ~/.zshrc:"
echo "  export GOOGLE_API_KEY=\"your-actual-api-key\""
echo ""
echo "Option 3: Set for current session only"
echo "  Run: export GOOGLE_API_KEY=\"your-actual-api-key\""
echo ""
echo "🔗 Get your API key from: https://console.cloud.google.com/apis/credentials"
echo ""
echo "📁 Configuration files created:"
echo "  - .qwen/mcp.json (MCP server configuration)"
echo "  - .env.example (Environment variable template)"
echo "  - .env (Your local environment - edit this)"
echo ""
echo "After setting your API key, restart Claude Code to load the Stitch MCP server."
echo ""
