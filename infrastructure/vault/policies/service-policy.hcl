path "secret/data/flyfleet/*" {
  capabilities = ["read"]
}

path "transit/sign/firmware-signing-key" {
  capabilities = ["update"]
}

path "transit/verify/firmware-signing-key" {
  capabilities = ["update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}
