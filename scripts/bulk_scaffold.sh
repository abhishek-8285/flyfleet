#!/bin/bash
set -e

cd /home/abhishek/Desktop/flyfleet

# Get all typical services
pushd services > /dev/null
SERVICES=$(ls -1d *-service* | grep -v "legacy$" | grep -v "auth-service$" | grep -v "new$")
popd > /dev/null

for svc in $SERVICES; do
    echo "Scaffolding $svc into ${svc}-new..."
    
    # Generate new template in *-new
    go run scripts/flyfleet-gen/main.go service "${svc}-new"
    
    # Write the appropriate go.mod for the NEW template
    cat << MODEOF > services/${svc}-new/go.mod
module github.com/flyfleet/${svc}-new

go 1.24.0

require (
	github.com/fleet-platform/shared v0.0.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/fleet-platform/shared => ../../packages/common/go
MODEOF
    
    # Tidy dependencies for the new template
    (cd services/${svc}-new && go mod tidy)
done

echo "Bulk generation into *-new complete. Active services were strictly preserved."
