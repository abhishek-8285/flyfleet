#!/bin/bash
echo -e "\033]0;PostgreSQL\007"
echo "Connecting to PostgreSQL Native Service..."
echo "Tailing pg_log (if accessible) or launching psql..."
sleep 2
PGPASSWORD="${PGPASSWORD:-YOUR_DB_PASSWORD}" psql -U fleet_user -d fleet_db -h localhost
exec bash
