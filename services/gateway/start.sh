#!/bin/bash
# Quick start script for AI Gateway

cd /data/data/com.termux/files/home/flyfleet/services/gateway

# Optional: Set your Qwen API credentials
# export QWEN_API_KEY="your-key-here"
# export QWEN_API_URL="https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"

echo "Starting AI Gateway..."
go run main.go
