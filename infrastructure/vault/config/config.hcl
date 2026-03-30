# Vault Development Configuration
# This is for development only. For production, use proper Vault setup with TLS, storage backend, etc.

disable_mlock = true

ui = true

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1  # Disable TLS for dev only
}

storage "file" {
  path = "/vault/data"
}

# Enable Secret Engines (auto-mounted)
secrets {
  path = "secret"
}

# Enable KV v2 (versioned secrets)
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Audit logging
audit "file" {
  path = "/vault/logs"
}

# Telemetry
telemetry {
  prometheus_retention_time = "30s"
  disable_hostname = true
}
