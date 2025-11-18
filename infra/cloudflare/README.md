# Cloudflare DNS Management with Terraform

This directory contains Terraform configuration for managing DNS records for the `apt-bundle.org` domain in Cloudflare.

## Overview

This Terraform configuration manages the following DNS records:

- **Root domain** (`apt-bundle.org`) → CNAME to `apt-bundle.github.io` (proxied)
- **WWW subdomain** (`www.apt-bundle.org`) → CNAME to `apt-bundle.github.io` (proxied)
- **Repository subdomain** (`repo.apt-bundle.org`) → CNAME to `apt-bundle.github.io` (NOT proxied)

> **Important:** The `repo.apt-bundle.org` record is intentionally NOT proxied because APT package managers require direct DNS resolution and do not work properly with Cloudflare's proxy.

## Prerequisites

- Cloudflare account with access to the `apt-bundle.org` domain
- Terraform >= 1.5.0 (for local testing)
- GitHub repository access (for CI/CD)

## Setup Instructions

### Step 1: Get Cloudflare API Token

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Go to **My Profile** → **API Tokens** (or visit https://dash.cloudflare.com/profile/api-tokens)
3. Click **Create Token**
4. Use the **Edit zone DNS** template, or create a custom token with:
   - **Permissions:**
     - Zone → DNS → Edit
     - Zone → Zone → Read
   - **Zone Resources:**
     - Include → Specific zone → `apt-bundle.org`
5. Click **Continue to summary** → **Create Token**
6. **Copy the token immediately** (you won't be able to see it again)

### Step 2: Find Cloudflare Zone ID

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Select the `apt-bundle.org` domain
3. Scroll down to the **API** section in the right sidebar
4. Copy the **Zone ID** (32-character hexadecimal string)

Alternatively, you can use the Cloudflare API:
```bash
curl -X GET "https://api.cloudflare.com/client/v4/zones?name=apt-bundle.org" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

### Step 3: Configure Remote State Backend

Choose one of the following options:

#### Option A: Terraform Cloud (Recommended - Free Tier)

1. Sign up at [Terraform Cloud](https://app.terraform.io/)
2. Create a new organization (e.g., `apt-bundle`)
3. Create a new workspace:
   - **Workspace name:** `cloudflare-dns`
   - **Workspace type:** Version control workflow
   - **Connect to VCS:** GitHub → Select `apt-bundle/apt-bundle` repository
   - **Working directory:** `infra/cloudflare`
4. In the workspace settings, add environment variables:
   - `CLOUDFLARE_API_TOKEN` (sensitive)
   - `CLOUDFLARE_ZONE_ID` (sensitive)
5. Update `terraform.tf` to uncomment the Terraform Cloud backend:
   ```hcl
   cloud {
     organization = "apt-bundle"
     workspaces {
       name = "cloudflare-dns"
     }
   }
   ```
6. Generate a Terraform Cloud API token:
   - Go to **User Settings** → **Tokens**
   - Create a new token
   - Add it as `TF_API_TOKEN` secret in GitHub

#### Option B: Cloudflare R2 (Recommended for Security)

**See [r2-setup.md](r2-setup.md) for complete secure setup instructions.**

R2 provides a secure, cost-effective backend with protection against unauthorized access:

1. Follow the detailed guide in `r2-setup.md` which covers:
   - Creating a secure, non-obvious bucket name
   - Setting up scoped R2 API tokens
   - Configuring access controls
   - Cost protection strategies
2. The `terraform.tf` file is already configured for R2 backend
3. Add the following GitHub Secrets (see `r2-setup.md` for details):
   - `R2_ACCESS_KEY_ID` - R2 API token access key
   - `R2_SECRET_ACCESS_KEY` - R2 API token secret key
   - `R2_BUCKET_NAME` - Your R2 bucket name
   - `R2_ENDPOINT` - R2 endpoint URL (https://<account-id>.r2.cloudflarestorage.com)

**Security Features:**
- Private buckets by default (no public access)
- Scoped API tokens (bucket-specific permissions)
- Non-obvious bucket naming to prevent enumeration
- Built-in encryption at rest
- Cost monitoring and alerts

#### Option C: AWS S3 Backend

1. Create an S3 bucket for state storage
2. Configure bucket versioning and encryption
3. Create IAM credentials with read/write access to the bucket
4. Update `terraform.tf` backend configuration for S3
5. Add AWS credentials to GitHub Secrets:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

### Step 4: Add GitHub Secrets

1. Go to your GitHub repository: https://github.com/apt-bundle/apt-bundle
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** and add:

   | Secret Name | Value | Description |
   |------------|-------|-------------|
   | `CLOUDFLARE_API_TOKEN` | Your API token from Step 1 | Cloudflare API authentication |
   | `CLOUDFLARE_ZONE_ID` | Your Zone ID from Step 2 | Cloudflare Zone identifier |
   | `TF_API_TOKEN` | Terraform Cloud token (if using Terraform Cloud) | For Terraform Cloud authentication |
   | `R2_ACCESS_KEY_ID` | R2 API token access key (if using R2) | R2 authentication |
   | `R2_SECRET_ACCESS_KEY` | R2 API token secret key (if using R2) | R2 authentication |
   | `R2_BUCKET_NAME` | R2 bucket name (if using R2) | R2 bucket identifier |
   | `R2_ENDPOINT` | R2 endpoint URL (if using R2) | R2 endpoint (https://<account-id>.r2.cloudflarestorage.com) |

### Step 5: Import Existing DNS Records

The root domain and www subdomain already exist in Cloudflare. You need to import them into Terraform state.

#### Find Existing Record IDs

First, get the record IDs for existing records:

```bash
# Set your API token and Zone ID
export CLOUDFLARE_API_TOKEN="your-api-token"
export CLOUDFLARE_ZONE_ID="your-zone-id"

# Get all DNS records
curl -X GET "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  -H "Content-Type: application/json" | jq '.result[] | {id: .id, name: .name, type: .type, content: .content, proxied: .proxied}'
```

Look for records matching:
- Name: `apt-bundle.org` (or `@`) → Type: `CNAME` → Content: `apt-bundle.github.io`
- Name: `www.apt-bundle.org` (or `www`) → Type: `CNAME` → Content: `apt-bundle.github.io`

#### Import Commands

Run these commands from the `infra/cloudflare` directory:

```bash
cd infra/cloudflare

# Initialize Terraform (if not already done)
terraform init

# Import root domain record
# Replace RECORD_ID_ROOT with the actual record ID from above
terraform import cloudflare_record.root CLOUDFLARE_ZONE_ID/RECORD_ID_ROOT

# Import www subdomain record
# Replace RECORD_ID_WWW with the actual record ID from above
terraform import cloudflare_record.www CLOUDFLARE_ZONE_ID/RECORD_ID_WWW
```

**Example:**
```bash
# If your Zone ID is abc123... and root record ID is def456...
terraform import cloudflare_record.root abc123def45678901234567890123456/def45678901234567890123456789012

# If www record ID is ghi789...
terraform import cloudflare_record.www abc123def45678901234567890123456/ghi78901234567890123456789012345
```

#### Verify Import

After importing, run a plan to verify:

```bash
terraform plan -var="cloudflare_zone_id=$CLOUDFLARE_ZONE_ID"
```

The plan should show:
- Root and www records: No changes (already match)
- Repo record: Will be created (new)

If there are differences, Terraform will show them. You may need to adjust the configuration to match existing settings.

### Step 6: Apply Configuration

#### Local Testing (Optional)

For local testing before pushing to GitHub:

```bash
cd infra/cloudflare

# Copy example variables file
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your Zone ID
# (This file is gitignored, so it won't be committed)

# Initialize
terraform init

# Plan
terraform plan

# Apply (if plan looks good)
terraform apply
```

#### Via GitHub Actions (Recommended)

1. Commit and push your changes:
   ```bash
   git add infra/cloudflare/
   git add .github/workflows/
   git commit -m "Add Terraform configuration for Cloudflare DNS management"
   git push origin main
   ```

2. The `terraform-apply.yml` workflow will automatically:
   - Run `terraform fmt` and `terraform validate`
   - Run `terraform plan`
   - Run `terraform apply` if plan succeeds

3. For pull requests, the `terraform-plan.yml` workflow will:
   - Run validation checks
   - Post a comment with the plan output

## Testing the Setup

### Verify DNS Records

After applying, verify the DNS records:

```bash
# Check root domain
dig apt-bundle.org CNAME +short

# Check www subdomain
dig www.apt-bundle.org CNAME +short

# Check repo subdomain
dig repo.apt-bundle.org CNAME +short
```

All should resolve to `apt-bundle.github.io`.

### Verify Proxied Status

```bash
# Root and www should be proxied (Cloudflare IPs)
dig apt-bundle.org +short
dig www.apt-bundle.org +short

# Repo should NOT be proxied (direct GitHub Pages IPs)
dig repo.apt-bundle.org +short
```

### Test APT Repository Access

The `repo.apt-bundle.org` subdomain should be accessible directly (not proxied):

```bash
curl -I https://repo.apt-bundle.org
```

This should return GitHub Pages headers, not Cloudflare headers.

## Workflow Details

### Terraform Plan Workflow

**Trigger:** Pull requests that modify files in `infra/cloudflare/`

**Actions:**
1. Checks Terraform formatting
2. Initializes Terraform
3. Validates configuration
4. Runs `terraform plan`
5. Posts plan output as PR comment

### Terraform Apply Workflow

**Trigger:** Pushes to `main` branch that modify files in `infra/cloudflare/`

**Actions:**
1. Checks Terraform formatting
2. Initializes Terraform
3. Validates configuration
4. Runs `terraform plan`
5. Runs `terraform apply` (auto-approve)
6. Posts summary with outputs

## Troubleshooting

### Import Errors

If import fails with "resource already managed":
- Check if records are already in state: `terraform state list`
- Remove from state if needed: `terraform state rm cloudflare_record.root`

### Plan Shows Changes for Existing Records

If plan shows changes for root/www records after import:
- Verify record settings match exactly (proxied status, TTL, etc.)
- Check if comments differ (Terraform may add comments)
- Run `terraform show` to see current state

### GitHub Actions Failures

**"CLOUDFLARE_API_TOKEN not set":**
- Verify secret is added in GitHub repository settings
- Check secret name matches exactly (case-sensitive)

**"Zone ID validation failed":**
- Verify `CLOUDFLARE_ZONE_ID` secret is a 32-character hex string
- Check for extra spaces or newlines

**"Backend initialization failed":**
- If using Terraform Cloud: Verify `TF_API_TOKEN` secret is set
- If using S3: Verify AWS credentials are set correctly

### DNS Propagation

After changes:
- DNS changes typically propagate within minutes
- Use `dig` or `nslookup` to verify from different locations
- Cloudflare proxied records may cache longer

## File Structure

```
infra/cloudflare/
├── README.md                 # This file
├── terraform.tf              # Provider and backend configuration
├── variables.tf              # Input variables
├── main.tf                   # DNS record resources
├── outputs.tf                # Output values
└── terraform.tfvars.example # Example variables file
```

## Security Notes

- **Never commit** `terraform.tfvars` (it's gitignored)
- **Never commit** `.tfstate` files (they may contain sensitive data)
- API tokens and Zone IDs are stored as GitHub Secrets
- Terraform state is stored remotely (Terraform Cloud or S3)
- Use least-privilege API tokens (zone-specific, DNS-only permissions)

## Additional Resources

- [Cloudflare Terraform Provider Documentation](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- [Terraform Cloud Documentation](https://www.terraform.io/docs/cloud)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review Terraform and Cloudflare provider documentation
3. Open an issue in the repository

