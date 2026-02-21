#!/bin/bash
set -e

# Configuration
OUT_DIR="./certs"
CA_KEY="${OUT_DIR}/ca.key"
CA_CERT="${OUT_DIR}/ca.crt"

# Service configuration
SERVICES=("auth-service" "tracking-service" "fleet-core-service-legacy" "billing-service" "web-bff" "maintenance-service" "alert-service" "admin-support-service")

mkdir -p "$OUT_DIR"

echo "Generating Root CA..."
openssl genrsa -out "$CA_KEY" 4096
openssl req -x509 -new -nodes -key "$CA_KEY" -sha256 -days 3650 -out "$CA_CERT" \
  -subj "/C=US/ST=CA/O=FlyFleet/OU=Root/CN=FlyFleet-Root-CA"

echo "Root CA generated."

for SVC in "${SERVICES[@]}"; do
  echo "Generating certificates for $SVC..."
  
  SVC_DIR="${OUT_DIR}/${SVC}"
  mkdir -p "$SVC_DIR"
  
  SVC_KEY="${SVC_DIR}/server.key"
  SVC_CSR="${SVC_DIR}/server.csr"
  SVC_CERT="${SVC_DIR}/server.crt"
  
  # Generate private key
  openssl genrsa -out "$SVC_KEY" 2048
  
  # Generate CSR
  openssl req -new -key "$SVC_KEY" -out "$SVC_CSR" \
    -subj "/C=US/ST=CA/O=FlyFleet/OU=Services/CN=${SVC}"

  # Generate Extension File for SAN
  EXT_FILE="${SVC_DIR}/extfile.cnf"
  cat > "$EXT_FILE" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SVC}
DNS.2 = localhost
DNS.3 = ${SVC}.local
IP.1 = 127.0.0.1
EOF

  # Sign Certificate
  openssl x509 -req -in "$SVC_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
    -out "$SVC_CERT" -days 365 -sha256 -extfile "$EXT_FILE"
    
  echo "Certificates for $SVC generated in $SVC_DIR"
  
  # Cleanup intermediate files
  rm "$SVC_CSR" "$EXT_FILE"
done

echo "Done. All certificates generated in $OUT_DIR."
