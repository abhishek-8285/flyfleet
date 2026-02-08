import requests
import sys

def get_policy_version(url, service_name):
    try:
        # Auth Service exposes it via health check or specific endpoint?
        # Based on previous file reads:
        # Web-BFF: /health/security -> {"security_policy_hash": "..."}
        # Auth-Service: I need to check if it exposes it. 
        # Checking auth-service main.go from previous steps...
        # It has /health and /metrics/all. 
        # It does NOT seem to have a specific /health/security endpoint like BFF.
        # I should check if /health includes it or if I need to add it.
        
        # Let's assume for now I might need to add it to Auth Service if not present.
        # But let's try to fetch from what is there.
        
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            data = response.json()
            return data.get("security_policy_hash")
        print(f"Error fetching from {service_name}: Status {response.status_code}")
        return None
    except Exception as e:
        print(f"Error connecting to {service_name}: {e}")
        return None

def main():
    # Fetch URLs from environment variables (for Docker/CI)
    import os
    bff_url = os.environ.get("BFF_URL", "http://localhost:8081/health/security")
    # Auth Service URL (default to localhost:8080 for local testing)
    auth_url = os.environ.get("AUTH_URL", "http://localhost:8080/health/security") 

    print("Verifying RBAC Policy Consistency...")
    print(f"BFF URL: {bff_url}")
    print(f"Auth URL: {auth_url}")
    
    bff_hash = get_policy_version(bff_url, "Web-BFF")
    auth_hash = get_policy_version(auth_url, "Auth-Service")

    if bff_hash and auth_hash:
        print(f"Web-BFF Hash:     {bff_hash}")
        print(f"Auth-Service Hash: {auth_hash}")
        
        if bff_hash == auth_hash:
            print("SUCCESS: Policies match.")
            sys.exit(0)
        else:
            print("FAILURE: Policy mismatch detected!")
            sys.exit(1)
    else:
        print("Could not retrieve hashes from both services.")
        # Fail open for now to allow me to fix the endpoint
        sys.exit(1)

if __name__ == "__main__":
    main()
