#!/bin/bash
echo -e "\033]0;FlyFleet Utils\007"
echo "Waiting for database to initialize... (4s)"
sleep 4
export DB_HOST=localhost
export DB_PORT=5432
export DB_PASSWORD=fleet_password
chmod +x migrate.sh
./migrate.sh
echo -e "\n\n🐘 Migrations Done!\n"
echo "👉 Run psql queries here or interact with the system!"
exec bash
