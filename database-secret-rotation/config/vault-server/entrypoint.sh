#! /bin/sh

set -e

export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_FORMAT="json"

# Spawn a new process for the development Vault server and wait for it to come online
# ref: https://www.vaultproject.io/docs/concepts/dev-server
vault server -dev -dev-listen-address="0.0.0.0:8200" &
sleep 5s

# Authenticate container's local Vault CLI
# ref: https://www.vaultproject.io/docs/commands/login
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

vault policy write trusted-orchestrator-policy /vault/config/truested-orchestrator-policy.hcl
vault policy write dev-policy /vault/config/dev-policy.hcl

# Enable AppRole auth method utilized by our web application
# ref: https://www.vaultproject.io/docs/auth/approle
vault auth enable approle

# Configure a specific AppRole role with associated parameters
# ref: https://www.vaultproject.io/api/auth/approle#parameters
#
# NOTE: we use artificially low ttl values to demonstrate the credential renewal logic
vault write auth/approle/role/dev-role \
    token_policies=dev-policy \
    secret_id_ttl="2m" \
    token_ttl="1m"  \
    token_max_ttl="6m"

# Overwrite our role id with a known value to simplify our demo
vault write auth/approle/role/dev-role/role-id role_id="${APPROLE_ROLE_ID}"

# Configure a token with permissions to act as a trusted orchestrator. For
# simplicity, we don't handle renewals in our simulated trusted orchestrator
# so we've set the ttl to a very long duration (768h). When this expires
# the web app will no longer receive a secret id and subsequently fail on the
# next attempted AppRole login.
# ref: https://www.vaultproject.io/docs/commands/token/create
vault token create \
    -id="${ORCHESTRATOR_TOKEN}" \
    -policy=trusted-orchestrator-policy \
    -ttl="768h"

#####################################
########## STATIC SECRETS ###########
#####################################

# Enable the kv-v2 secrets engine, passing in the path parameter
# ref: https://www.vaultproject.io/docs/secrets/kv/kv-v2
vault secrets enable -path=kv-v2 kv-v2

# Seed the kv-v2 store with an entry our web app will use
vault kv put "${API_KEY_PATH}" "${API_KEY_FIELD}=my-secret-key"


#####################################
########## DYNAMIC SECRETS ##########
#####################################

# Enable the database secrets engine
# ref: https://www.vaultproject.io/docs/secrets/databases
vault secrets enable database

vault write /database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="dev-only" \
    connection_url="postgresql://{{username}}:{{password}}@${DATABASE_HOSTNAME}:${DATABASE_PORT}/postgres?sslmode=disable" \
    username="vault_db_user" \
    password="vault_db_password"

# Rotate the password for 'vault_db_user', ensures the user is only accessible
# by Vault itself
vault write -force database/config/my-postgresql-database

vault write database/roles/dev-readonly \
    db_name=my-postgresql-datatbase \
    create_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT readonly TO \"{{name}}\";" \
    renew_statements="ALTER ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT readonly TO \"{{name}}\";" \
    default_ttl="100s" \
    max_ttl="300s"

# This container is now healthy
touch /tmp/healthy

# Keep container alive
tail -f /dev/null & trap 'kill %1' TERM ; wait

