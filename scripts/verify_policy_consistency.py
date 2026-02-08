import requests
import sys
import json

def verify_policy_consistency(endpoints):
    hashes = {}
    failed = False

    print("--- Security Policy Consistency Check ---")
    
    for service, url in endpoints.items():
        try:
            response = requests.get(url, timeout=5)
            response.raise_for_status()
            data = response.json()
            policy_hash = data.get("security_policy_hash")
            
            if not policy_hash:
                print(f"[FAIL] {service}: Policy hash missing in response from {url}")
                failed = True
                continue
                
            hashes[service] = policy_hash
            print(f"[OK]   {service}: {policy_hash}")
            
        except Exception as e:
            print(f"[FAIL] {service}: Could not connect to {url} - {str(e)}")
            failed = True

    if failed:
        print("\nCRITICAL: One or more services failed to report their security policy.")
        sys.exit(1)

    # Validate equality
    unique_hashes = set(hashes.values())
    if len(unique_hashes) > 1:
        print("\n[!] ERROR: Security Policy Mismatch Detected!")
        print("Versions found:")
        for service, h in hashes.items():
            print(f"  - {service}: {h}")
        print("\nPotential Risk: Services may be enforcing different RBAC rules, leading to privilege bypass or unexpected denials.")
        sys.exit(1)
    
    print("\n[SUCCESS] All services are running identical security policies (Version: {}).".format(list(unique_hashes)[0]))
    sys.exit(0)

if __name__ == "__main__":
    # In a real environment, these would be loaded from service discovery (e.g., Consul)
    # or passed as environment variables.
    TARGETS = {
        "auth-service": "http://localhost:8080/health",
        "web-bff":      "http://localhost:8081/health/security"
    }
    
    # Optional: Override targets from args if needed
    if len(sys.argv) > 1:
        try:
            TARGETS = json.loads(sys.argv[1])
        except:
            print("Usage: python verify_policy_consistency.py '[{\"name\": \"url\"}]'")

    verify_policy_consistency(TARGETS)
