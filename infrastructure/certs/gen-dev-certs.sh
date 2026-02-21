#!/usr/bin/env bash
set -euo pipefail

# Generate self-signed CA and service certificates for local mTLS development.
# Usage: ./gen-dev-certs.sh [output_dir]

CERT_DIR="${1:-./certs}"
DAYS=365
CA_SUBJECT="/CN=FlyFleet Dev CA/O=FlyFleet/OU=Engineering"

SERVICES=(
  "auth-service"
  "fleet-core-service"
  "tracking-service"
  "web-bff"
)

echo "🔐 Generating FlyFleet dev certificates in ${CERT_DIR}..."
mkdir -p "${CERT_DIR}"

# ── 1. Generate CA ──────────────────────────────────
if [ ! -f "${CERT_DIR}/ca.key" ]; then
  echo "  → Generating CA key and certificate..."
  openssl genrsa -out "${CERT_DIR}/ca.key" 4096
  openssl req -x509 -new -nodes \
    -key "${CERT_DIR}/ca.key" \
    -sha256 -days ${DAYS} \
    -out "${CERT_DIR}/ca.crt" \
    -subj "${CA_SUBJECT}"
else
  echo "  → CA already exists, reusing..."
fi

# ── 2. Generate service certs ───────────────────────
for SVC in "${SERVICES[@]}"; do
  echo "  → Generating cert for ${SVC}..."
  
  # Create CSR config with SAN
  cat > "${CERT_DIR}/${SVC}.cnf" <<EOF
[req]
default_bits = 2048
prompt = no
distinguished_name = dn
req_extensions = v3_req

[dn]
CN = ${SVC}
O = FlyFleet
OU = Engineering

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SVC}
DNS.2 = ${SVC}.flyfleet.svc.cluster.local
DNS.3 = localhost
IP.1 = 127.0.0.1
EOF

  # Generate key + CSR
  openssl genrsa -out "${CERT_DIR}/${SVC}.key" 2048
  openssl req -new \
    -key "${CERT_DIR}/${SVC}.key" \
    -out "${CERT_DIR}/${SVC}.csr" \
    -config "${CERT_DIR}/${SVC}.cnf"

  # Sign with CA
  openssl x509 -req \
    -in "${CERT_DIR}/${SVC}.csr" \
    -CA "${CERT_DIR}/ca.crt" \
    -CAkey "${CERT_DIR}/ca.key" \
    -CAcreateserial \
    -out "${CERT_DIR}/${SVC}.crt" \
    -days ${DAYS} \
    -sha256 \
    -extensions v3_req \
    -extfile "${CERT_DIR}/${SVC}.cnf"

  # Clean up CSR and config
  rm -f "${CERT_DIR}/${SVC}.csr" "${CERT_DIR}/${SVC}.cnf"
done

echo ""
echo "✅ Certificates generated in ${CERT_DIR}/"
echo ""
echo "Files:"
ls -la "${CERT_DIR}/"
echo ""
echo "Usage in docker-compose.yml:"
echo "  volumes:"
echo "    - ./certs:/certs:ro"
echo "  environment:"
echo '    - TLS_CERT_FILE=/certs/${SERVICE_NAME}.crt'
echo '    - TLS_KEY_FILE=/certs/${SERVICE_NAME}.key'
echo '    - TLS_CA_FILE=/certs/ca.crt'
