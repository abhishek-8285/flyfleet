#!/bin/bash
echo "--- DIAGNOSTICS ---" > diagnostics.txt
date >> diagnostics.txt

echo "Checking Postgres (5432)..." >> diagnostics.txt
lsof -i :5432 >> diagnostics.txt 2>&1 || echo "Port 5432 not listening" >> diagnostics.txt

echo "Checking Redis (6379)..." >> diagnostics.txt
lsof -i :6379 >> diagnostics.txt 2>&1 || echo "Port 6379 not listening" >> diagnostics.txt

echo "Checking Redis (6380)..." >> diagnostics.txt
lsof -i :6380 >> diagnostics.txt 2>&1 || echo "Port 6380 not listening" >> diagnostics.txt

echo "Checking NATS (4222)..." >> diagnostics.txt
lsof -i :4222 >> diagnostics.txt 2>&1 || echo "Port 4222 not listening" >> diagnostics.txt

echo "Checking Backend (8084)..." >> diagnostics.txt
lsof -i :8084 >> diagnostics.txt 2>&1 || echo "Port 8084 not listening" >> diagnostics.txt

echo "Checking Frontend (3000)..." >> diagnostics.txt
lsof -i :3000 >> diagnostics.txt 2>&1 || echo "Port 3000 not listening" >> diagnostics.txt

echo "Processes:" >> diagnostics.txt
ps aux | grep -E "postgres|redis|nats|fleet-core|vite|node" >> diagnostics.txt

echo "--- END ---" >> diagnostics.txt
