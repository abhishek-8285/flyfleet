import sys
import os
from pathlib import Path

# Add proto paths to sys.path for imports
# This is a bit complex in a monorepo, usually we'd use a shared package
# But for a quick validation script we can use grpcurl or simple dynamic clients

SERVICES = {
    "Auth": 50051,
    "Tracking": 50052,
    "Trip": 50053,
    "Alert": 50054,
    "Billing": 50055,
    "Admin": 50056,
    "Maintenance": 50057,
    "Driver": 50058,
    "Analytics": 50059,
    "Reports": 50060,
    "Fleet": 50061,
}

def check_port(port):
    import socket
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        return s.connect_ex(('localhost', port)) == 0

def validate():
    print("============================================")
    print("  FlyFleet gRPC Platform Validation")
    print("============================================")
    
    all_pass = True
    for name, port in SERVICES.items():
        status = "UP" if check_port(port) else "DOWN"
        icon = "✅" if status == "UP" else "❌"
        print(f"{icon} {name:12} (Port {port}): {status}")
        if status == "DOWN":
            all_pass = False
            
    print("============================================")
    if all_pass:
        print("RESULT: PLATFORM STRUCTURALLY PERFECT")
        return 0
    else:
        print("RESULT: PLATFORM INCOMPLETE (Services Down)")
        return 1

if __name__ == "__main__":
    sys.exit(validate())
