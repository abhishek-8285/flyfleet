#!/bin/bash
echo -e "\033]0;Redis Cache\007"
echo "Connecting to Redis Server..."
redis-cli ping
exec bash
