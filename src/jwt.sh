#!/bin/bash

set -e

print_usage() {
  echo "Usage: $0 [-k private_key_path] [--priv-key private_key_path] [--ttl duration]"
  echo "  -k, --priv-key private_key_path  Path to the private key file (default: ./private_key.pem)"
  echo "  --ttl          duration          Duration in seconds for which the token will be valid (default: 600)"
}

# Function to URL-safe base64 encode
base64_url_encode() {
  openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

# Default values
KEY_PATH="./private_key.pem"
TTL="600"

# Parse command line options
while [[ "$#" -gt 0 ]]; do
  case $1 in
  -k | --priv-key)
    KEY_PATH="$2"
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

# Check if the private key file exists
if [ ! -f "$KEY_PATH" ]; then
  echo "Private key file not found: $KEY_PATH"
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

payload=$(
  cat <<EOF
{
	"iss": "space-invaders",
	"sub": "sarumaj",
	"aud": ["space-invaders"],
	"iat": $iat,
	"exp": $((iat + $TTL))
}
EOF
)

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
