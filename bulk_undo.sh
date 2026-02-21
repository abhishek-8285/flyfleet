#!/bin/bash
set -e

SERVICES=$(ls -1 services/ | grep "-service-legacy$")

cd /home/abhishek/Desktop/flyfleet

for legacy_svc in $SERVICES; do
    svc=${legacy_svc%-legacy}
    
    if [ "$svc" == "auth-service" ]; then
        continue # Auth is refactored properly
    fi
    
    echo "Restoring $svc..."
    
    # Remove the empty new template
    rm -rf services/$svc
    
    # Restore the legacy code back to its original name
    mv services/$legacy_svc services/$svc
done
