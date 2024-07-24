#!/bin/bash

set -e

print_usage() {
  echo "Usage: $0 [-d target_directory] [--directory target_directory]"
  echo "  -d, --directory target_directory   Directory where the project will be created (default: current directory)"
}

# Function to URL-safe base64 encode
base64_url_encode() {
  openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

# Default values
KEY_PATH="."

# Parse command line options
while [[ "$#" -gt 0 ]]; do
  case $1 in
  -k | --key)
    KEY_PATH="$2"
    shift 2
    ;;
  *)
    echo "Unknown option: $1"
    print_usage
    exit 1
    ;;
  esac
done

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
	"exp": $((iat + 600))
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
