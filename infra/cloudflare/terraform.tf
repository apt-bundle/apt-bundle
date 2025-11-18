terraform {
  required_version = ">= 1.5.0"

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }

  # ============================================================================
  # BACKEND: Where Terraform stores its state file
  # ============================================================================
  # This tells Terraform WHERE to store the state file (metadata about what
  # Terraform manages). This uses Cloudflare R2 (S3-compatible storage).
  #
  # Authentication: Uses R2 API tokens (separate from Cloudflare API tokens)
  #   - R2_ACCESS_KEY_ID (GitHub Secret)
  #   - R2_SECRET_ACCESS_KEY (GitHub Secret)
  #   - R2_BUCKET_NAME (GitHub Secret)
  #   - R2_ENDPOINT (GitHub Secret)
  #
  # See r2-setup.md for secure setup instructions
  # Backend configuration is provided entirely via backend.hcl
  # The backend.hcl file is generated automatically in GitHub Actions from secrets
  # For local development, create backend.hcl manually (it's gitignored)
  backend "s3" {
  }
}

# ============================================================================
# PROVIDER: How Terraform manages Cloudflare resources (DNS records)
# ============================================================================
# This tells Terraform HOW to manage Cloudflare resources (like DNS records).
# The provider uses Cloudflare's API to create/update/delete DNS records.
#
# Authentication: Uses Cloudflare API token (separate from R2 tokens)
#   - CLOUDFLARE_API_TOKEN (GitHub Secret)
#   - Permissions needed: Zone → DNS → Edit, Zone → Zone → Read
#
# This is completely separate from the backend above. The backend stores
# Terraform's state, while the provider manages your actual DNS records.
provider "cloudflare" {
  # API token should be provided via CLOUDFLARE_API_TOKEN environment variable
  # or via GitHub Actions secrets
}

