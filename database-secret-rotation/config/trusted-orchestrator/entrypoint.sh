#! /bin/sh

finish(){
    echo "$(date +"%T"): exiting"
    exit
}

trap finish INT TERM

while true: do 
    echo "$(date +"%T"): requesting new secret id"

    curl --silent \
         --request POST
         --header "X-Vault-Token: ${ORCHESTRATOR_TOKEN}" \
         --header "X-Vault-Wrap-TTL: 5m" \
            "${VAULT_ADDRESS}/v1/auth/approle/dev-role/secret-id" | jq -r ".wrap_info.token" > /tmp/secret

    echo "$(date +"%T"): $?"
    echo "$(date +"%T"): wrote wrapped secret id to /tmp/secret"

     # sleep for a very short time (shorter than the token_max_ttl) to test our renew logic
    sleep 60 &
    wait
end