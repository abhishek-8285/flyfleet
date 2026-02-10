#!/usr/bin/env python3

# Quick fix script for the most critical hardcoded values

import os
import re

# Critical files to fix (highest priority)
critical_files = [
    ("/home/abhishek/Desktop/flyfleet/services/auth-service/internal/config/config.go", [
        ('"postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"', '"$DATABASE_URL"'),
        ('"your-super-secret-jwt-key-change-in-production"', '"$JWT_SECRET"'),
        ('"http://localhost:8083/api/v1"', '"$DRIVER_SERVICE_URL"'),
        ('"http://localhost:8084/api/v1"', '"$VEHICLE_SERVICE_URL"'),
        ('"http://localhost:8082/api/v1"', '"$TRIP_SERVICE_URL"'),
        ('"http://localhost:8085/api/v1"', '"$ALERT_SERVICE_URL"'),
    ]),
    # Add more critical files here
]

def apply_fixes():
    print("Applying fixes for hardcoded values...")
    
    for filepath, replacements in critical_files:
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
                print(f"✓ Fixed: {filepath}")
            else:
                print(f"- No changes needed: {filepath}")
                
        except Exception as e:
            print(f"✗ Error fixing {filepath}: {e}")

if __name__ == "__main__":
    apply_fixes()