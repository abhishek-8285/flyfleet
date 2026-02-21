#!/bin/bash
set -e

SERVICES=$(ls -1 services/ | grep "-service$" | grep -v "legacy$")

cd /home/abhishek/Desktop/flyfleet

for svc in $SERVICES; do
    if [ "$svc" == "auth-service" ]; then
        continue # Already done
    fi
    
    echo "Refactoring $svc..."
    
    # Generate new template
    go run scripts/flyfleet-gen/main.go service "${svc}-new"
    
    # Rename directories
    mv services/$svc services/${svc}-legacy
    mv services/${svc}-new services/$svc
    
    # Fix go.mod dependency
    cat << MODEOF > services/$svc/go.mod
module github.com/flyfleet/$svc

go 1.24.0

require (
github.com/fleet-platform/shared v0.0.0
google.golang.org/grpc v1.72.0
google.golang.org/protobuf v1.36.10
)

replace github.com/fleet-platform/shared => ../../packages/common/go
MODEOF
    
    # Tidy modules
    (cd services/$svc && go mod tidy)
done
