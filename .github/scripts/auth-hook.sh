#!/bin/bash

# Set variables from the environment
DOMAIN="${CERTBOT_DOMAIN}"
VALIDATION="${CERTBOT_VALIDATION}"
NAMECHEAP_API_USER="${NAMECHEAP_API_USER}"
NAMECHEAP_API_KEY="${NAMECHEAP_API_KEY}"
NAMECHEAP_USERNAME="${NAMECHEAP_USERNAME}"
NAMECHEAP_CLIENT_IP="${NAMECHEAP_CLIENT_IP}"

# Extract SLD and TLD
SLD=$(echo $DOMAIN | rev | cut -d'.' -f2 | rev)
TLD=$(echo $DOMAIN | rev | cut -d'.' -f1 | rev)

# Create the DNS record
RESPONSE=$(curl -s \
    "https://api.namecheap.com/xml.response" \
    --data-urlencode "apiuser=${NAMECHEAP_API_USER}" \
    --data-urlencode "apikey=${NAMECHEAP_API_KEY}" \
    --data-urlencode "username=${NAMECHEAP_USERNAME}" \
    --data-urlencode "Command=namecheap.domains.dns.setHosts" \
    --data-urlencode "ClientIp=${NAMECHEAP_CLIENT_IP}" \
    --data-urlencode "SLD=${SLD}" \
    --data-urlencode "TLD=${TLD}" \
    --data-urlencode "HostName1=_acme-challenge" \
    --data-urlencode "RecordType1=TXT" \
    --data-urlencode "Address1=${VALIDATION}" \
    --data-urlencode "TTL1=120")

# Output the response for debugging
echo "Auth Hook Response: $RESPONSE"

# Check if the response contains an error
ERROR_STATUS=$(echo "$RESPONSE" | xmllint --xpath 'string(//ApiResponse/@Status)' -)
ERROR_MESSAGE=$(echo "$RESPONSE" | xmllint --xpath 'string(//Error)' -)

if [[ "$ERROR_STATUS" == "ERROR" ]]; then
    echo "Error in API response: $ERROR_MESSAGE"
    exit 1
fi

# Wait for DNS propagation
PROPAGATION_DELAY=30 # seconds
MAX_ATTEMPTS=10
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    echo "Waiting for DNS propagation..."
    sleep $PROPAGATION_DELAY
    ATTEMPT=$((ATTEMPT + 1))
    dig_txt_record=$(dig +short TXT _acme-challenge.${DOMAIN})

    if [[ "$dig_txt_record" == "\"$VALIDATION\"" ]]; then
        echo "DNS TXT record has propagated: $dig_txt_record"
        exit 0
    fi
done

echo "DNS TXT record did not propagate in time."
exit 1
