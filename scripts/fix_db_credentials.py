#!/usr/bin/env python3

# Fix hardcoded database credentials in service main.go files

import os
import re

# Service files with hardcoded database credentials
service_configs = [
    # Alert Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/alert-service/cmd/main.go", [
        ('"postgres://fleet:fleet123@localhost:5432/fleetdb"', '""'),
    ]),
    
    # Analytics Service  
    ("/home/abhishek/Desktop/flyfleet/services/microservices/analytics-service/cmd/main.go", [
        ('"postgres://user:password@localhost:5432/analytics_db?sslmode=disable"', '""'),
    ]),
    
    # Billing Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/billing-service/cmd/main.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '""'),
    ]),
    
    # Driver Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/driver-service/cmd/main.go", [
        ('"postgres://fleet:fleet123@localhost:5432/fleetdb"', '""'),
    ]),
    
    # DR Service
    ("/home/abhishek/Desktop/flyfleet/services/dr-service/cmd/main.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '""'),
    ]),
    
    # Geofencing Service
    ("/home/abhishek/Desktop/flyfleet/services/geofencing-service/cmd/main.go", [
        ('"postgres://fleet:fleet@localhost:5433/fleetdb"', '""'),
    ]),
    
    # Governance Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/governance-service/cmd/main.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '""'),
    ]),
    
    # Maintenance Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/maintenance-service/cmd/main.go", [
        ('"postgres://fleet:fleet123@localhost:5432/fleetdb"', '""'),
    ]),
    
    # Notification Service
    ("/home/abhishek/Desktop/flyfleet/services/notification-service/cmd/main.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '""'),
    ]),
    
    # Reports Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/reports-service/cmd/main.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '""'),
    ]),
    
    # Route Optimization Service
    ("/home/abhishek/Desktop/flyfleet/services/route-optimization-service/cmd/main.go", [
        ('"postgres://fleet:fleet@localhost:5433/fleetdb"', '""'),
    ]),
    
    # Tracking Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/tracking-service/cmd/main.go", [
        ('"postgres://fleet:fleet@localhost:5433/fleetdb"', '""'),
    ]),
    
    # Trip Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/trip-service/cmd/main.go", [
        ('"postgres://fleet:fleet123@localhost:5432/fleetdb"', '""'),
    ]),
    
    # Vehicle Service
    ("/home/abhishek/Desktop/flyfleet/services/microservices/vehicle-service/cmd/main.go", [
        ('"postgres://fleet:fleet@localhost:5433/fleetdb"', '""'),
    ]),
    
    # Video Service
    ("/home/abhishek/Desktop/flyfleet/services/video-service/cmd/main.go", [
        ('"postgres://fleet:fleet@localhost:5433/fleetdb"', '""'),
    ]),
]

def fix_database_credentials():
    print("Fixing hardcoded database credentials...")
    fixed_count = 0
    
    for filepath, replacements in service_configs:
        if not os.path.exists(filepath):
            print(f"File not found: {filepath}")
            continue
            
        try:
            with open(filepath, 'r') as f:
                content = f.read()
            
            original_content = content
            
            for old, new in replacements:
                content = content.replace(old, new)
            
            if content != original_content:
                with open(filepath, 'w') as f:
                    f.write(content)
                print(f"✓ Fixed database credentials in: {filepath}")
                fixed_count += 1
            else:
                print(f"- No changes needed: {filepath}")
                
        except Exception as e:
            print(f"✗ Error fixing {filepath}: {e}")
    
    print(f"\nFixed {fixed_count} service configuration files")

if __name__ == "__main__":
    fix_database_credentials()