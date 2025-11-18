#!/bin/bash
# Helper script to import existing Cloudflare DNS records into Terraform state
#
# Usage:
#   export CLOUDFLARE_API_TOKEN="your-token"
#   export CLOUDFLARE_ZONE_ID="your-zone-id"
#   ./import-records.sh

set -euo pipefail

if [ -z "${CLOUDFLARE_API_TOKEN:-}" ]; then
    echo "Error: CLOUDFLARE_API_TOKEN environment variable is not set"
    exit 1
fi

if [ -z "${CLOUDFLARE_ZONE_ID:-}" ]; then
    echo "Error: CLOUDFLARE_ZONE_ID environment variable is not set"
    exit 1
fi

# Check if jq is installed (required for parsing JSON)
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Install it with: apt-get install jq"
    exit 1
fi

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "Error: terraform is required but not installed. Install it from https://www.terraform.io/downloads"
    exit 1
fi

echo "Fetching DNS records from Cloudflare..."
RECORDS=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  -H "Content-Type: application/json")

# Find root domain record (apt-bundle.org or @)
ROOT_RECORD_ID=$(echo "$RECORDS" | jq -r '.result[] | select(.name == "apt-bundle.org" or .name == "@") | select(.type == "CNAME") | select(.content == "apt-bundle.github.io") | .id' | head -n1)

# Find www subdomain record
WWW_RECORD_ID=$(echo "$RECORDS" | jq -r '.result[] | select(.name == "www.apt-bundle.org" or .name == "www") | select(.type == "CNAME") | select(.content == "apt-bundle.github.io") | .id' | head -n1)

if [ -z "$ROOT_RECORD_ID" ]; then
    echo "Error: Could not find root domain CNAME record"
    exit 1
fi

if [ -z "$WWW_RECORD_ID" ]; then
    echo "Error: Could not find www subdomain CNAME record"
    exit 1
fi

echo ""
echo "Found records:"
echo "  Root domain: $ROOT_RECORD_ID"
echo "  WWW subdomain: $WWW_RECORD_ID"
echo ""

# Check if terraform is initialized
if [ ! -d ".terraform" ]; then
    echo "Initializing Terraform..."
    terraform init
fi

echo "Importing root domain record..."
terraform import cloudflare_record.root "$CLOUDFLARE_ZONE_ID/$ROOT_RECORD_ID"

echo "Importing www subdomain record..."
terraform import cloudflare_record.www "$CLOUDFLARE_ZONE_ID/$WWW_RECORD_ID"

echo ""
echo "✅ Import complete!"
echo ""
echo "Run 'terraform plan -var=\"cloudflare_zone_id=$CLOUDFLARE_ZONE_ID\"' to verify."

