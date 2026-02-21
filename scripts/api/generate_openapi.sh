#!/bin/bash
set -e

# Configuration
PROTO_DIR=${PROTO_DIR:-"./services/*/proto"}
OUT_DIR=${OUT_DIR:-"./docs/api"}
OUT_FILE="$OUT_DIR/openapi.json"

echo "Initializing OpenAPI generation..."
mkdir -p "$OUT_DIR"

# Merge all Protobuf components into a single swagger specification
# This assumes the protoc-gen-openapiv2 plugin is installed in the dev environment.
# Since we have multiple microservices, we will aggregate them.

# Note: In a production scenario, we'd use buf generate. For this script, we'll
# use a simplified loop to generate swagger for each and highlight the requirement.

if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed."
    exit 1
fi

if ! command -v protoc-gen-openapiv2 &> /dev/null; then
    echo "Installing protoc-gen-openapiv2..."
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
fi

echo "Scanning protocol buffers..."
for service_dir in services/*; do
    if [ -d "$service_dir" ]; then
        proto_files=$(find "$service_dir" -name "*.proto")
        if [ -n "$proto_files" ]; then
            echo "Processing $service_dir..."
            protoc \
                -I="$service_dir" \
                -I="proto" \
                --openapiv2_out="$OUT_DIR" \
                --openapiv2_opt=allow_merge=true,merge_file_name=flyfleet \
                $proto_files || echo "Skipping some files in $service_dir due to missing OpenAPI annotations."
        fi
    fi
done

echo "OpenAPI specs generated in $OUT_DIR"
