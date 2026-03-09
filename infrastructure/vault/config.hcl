storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1
}

ui = true

# Key rotation and audit logging policies would be enabled here via CLI/API
# vault path-help secret
# vault secrets enable transit
