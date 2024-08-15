#!/bin/bash

set -e

print_usage() {
  echo "Usage: $0 [-k key_path] [--rsa-key key_path] [--ttl duration] [-s subject] [--subject subject]"
  echo "  -k, --rsa-key key_path   Path to the RSA key file (default: ./rsa.pem)"
  echo "  -s, --subject subject    Subject of the token (default: sarumaj)"
  echo "  --ttl duration           Duration in seconds for which the token will be valid (default: 600)"
}

# Function to URL-safe base64 encode
base64_url_encode() {
  openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

# Default values
KEY_PATH="./rsa.pem"
TTL="600"
SUBJECT="sarumaj"

# Parse command line options
while [[ "$#" -gt 0 ]]; do
  case $1 in
  -k | --rsa-key)
    KEY_PATH="$2"
    shift 2
    ;;
  -s | --subject)
    SUBJECT="$2"
    shift 2
    ;;
  --ttl)
    TTL="$2"
    shift 2
    ;;
  *)
    echo "Unknown option: $1"
    print_usage
    exit 1
    ;;
  esac
done

# Check if the RSA key file exists
if [ ! -f "$KEY_PATH" ]; then
  echo "RSA key file not found: $KEY_PATH"
  print_usage
  exit 1
fi

# Validate TTL
if ! [[ "$TTL" =~ ^[0-9]+$ ]]; then
  echo "Invalid TTL: $TTL"
  print_usage
  exit 1
fi

# Get the current timestamp for 'iat' claim
iat=$(date +%s)

# Define header and payload as variables
header='{
	"alg": "RS256",
	"typ": "JWT"
}'

payload=$(jq -n --arg iat "$iat" --arg ttl "$TTL" --arg sub "$SUBJECT" '
  .iat = ($iat | tonumber) |
  .iss = "space-invaders" |
  .sub = $sub |
  .aud = ["space-invaders"] |
  if ($ttl | tonumber) > 0 then
    .exp = (($iat | tonumber) + ($ttl | tonumber))
  else
    .
  end
')

# Base64 URL encode the header and payload
header_base64=$(echo -n "${header}" | base64_url_encode)
payload_base64=$(echo -n "${payload}" | base64_url_encode)

# Create unsigned token
unsigned_token="${header_base64}.${payload_base64}"

# Sign the token
signature=$(echo -n "${unsigned_token}" | openssl dgst -sha256 -sign "$KEY_PATH" | base64_url_encode)

# Combine to form the final JWT
jwt="${unsigned_token}.${signature}"
echo "${jwt}"
