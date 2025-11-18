# Secure R2 Backend Setup Guide

This guide explains how to securely configure Cloudflare R2 as a Terraform state backend with protection against unauthorized access and cost overruns.

## Security Strategy

### 1. **Non-Obvious Bucket Naming**
- Use a random or hashed bucket name instead of descriptive names
- Prevents easy discovery through enumeration
- Example: `tf-state-a7f3b2c9d1e4f5g6` instead of `apt-bundle-terraform-state`

### 2. **Minimal Permissions**
- Create R2 API tokens with **Object Read/Write** permissions only
- Scope tokens to specific bucket (not account-wide)
- Use separate tokens for different environments if needed

### 3. **Bucket Privacy**
- R2 buckets are private by default (no public access)
- Never enable public access
- No public URLs or signed URLs needed for state storage

### 4. **State Encryption**
- Terraform state files contain sensitive data
- R2 encrypts data at rest automatically
- Consider enabling additional encryption if required

### 5. **Access Monitoring**
- Monitor R2 usage via Cloudflare Dashboard
- Set up alerts for unusual activity
- Review access logs regularly

## Step-by-Step Setup

### Step 1: Create R2 Bucket

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Go to **R2** → **Create bucket**
3. **Bucket name**: Use a non-obvious name (see naming strategy below)
   - Recommended: Generate a random name like `tf-state-<random-hex>`
   - Example: `tf-state-a7f3b2c9d1e4f5g6`
4. **Location**: Choose closest to your primary region
5. **Public access**: **DISABLED** (default - keep it this way)
6. Click **Create bucket**

#### Generate Secure Bucket Name

```bash
# Generate a random bucket name
openssl rand -hex 8 | sed 's/^/tf-state-/'
# Example output: tf-state-a7f3b2c9d1e4f5g6
```

**Important**: Save this bucket name securely - you'll need it for configuration.

### Step 2: Create R2 API Token

1. In Cloudflare Dashboard, go to **R2** → **Manage R2 API Tokens**
2. Click **Create API token**
3. Configure token:
   - **Token name**: `terraform-state-r2` (descriptive for your reference)
   - **Permissions**: 
     - ✅ **Object Read**
     - ✅ **Object Write**
     - ❌ **Object Delete** (optional - only if you want Terraform to delete old state versions)
     - ❌ **Admin Read/Write** (NOT needed)
   - **TTL**: Set expiration if desired (or leave blank for no expiration)
   - **Bucket restriction**: Select your specific bucket (important!)
4. Click **Create API Token**
5. **Copy the token immediately** - you won't see it again
   - Access Key ID: `xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
   - Secret Access Key: `yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy`

### Step 3: Get Your Account ID

1. In Cloudflare Dashboard, select any domain
2. Scroll to **API** section in right sidebar
3. Copy your **Account ID** (32-character hex string)
4. Or go to **R2** → Your bucket → **Settings** → Account ID is shown there

### Step 4: Configure Terraform Backend

Update `terraform.tf` with your R2 configuration:

```hcl
backend "s3" {
  bucket     = "tf-state-a7f3b2c9d1e4f5g6"  # Your bucket name
  key        = "cloudflare/terraform.tfstate"
  region     = "auto"  # R2 uses "auto" region
  endpoints  = {
    s3 = "https://<account-id>.r2.cloudflarestorage.com"
  }
  access_key                  = ""  # Provided via AWS_ACCESS_KEY_ID env var
  secret_key                  = ""  # Provided via AWS_SECRET_ACCESS_KEY env var
  skip_credentials_validation = true
  skip_region_validation      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  force_path_style            = true
}
```

Replace `<account-id>` with your actual Account ID.

### Step 5: Add GitHub Secrets

Add these secrets to your GitHub repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add the following secrets:

   | Secret Name | Value | Description |
   |------------|-------|-------------|
   | `R2_ACCESS_KEY_ID` | Your R2 Access Key ID | R2 API token access key |
   | `R2_SECRET_ACCESS_KEY` | Your R2 Secret Access Key | R2 API token secret key |
   | `R2_BUCKET_NAME` | Your bucket name | R2 bucket name (e.g., `tf-state-a7f3b2c9d1e4f5g6`) |
   | `R2_ENDPOINT` | R2 endpoint URL | R2 endpoint (https://<account-id>.r2.cloudflarestorage.com) |

### Step 6: Update GitHub Actions Workflows

The workflows will automatically use these secrets. See `terraform.tf` for the backend configuration.

## Security Best Practices

### ✅ DO:
- Use non-obvious bucket names
- Scope R2 tokens to specific buckets only
- Store credentials in GitHub Secrets (never commit)
- Monitor R2 usage regularly
- Use separate tokens for different environments
- Set token expiration dates if possible
- Rotate tokens periodically

### ❌ DON'T:
- Use descriptive bucket names like "terraform-state" or "apt-bundle-state"
- Create account-wide R2 tokens
- Commit credentials or bucket names to git
- Enable public access on R2 buckets
- Share R2 tokens publicly
- Use the same token for multiple projects

## Cost Protection

### Understanding R2 Pricing
- **Storage**: $0.015 per GB/month (very cheap)
- **Class A Operations** (writes): $4.50 per million
- **Class B Operations** (reads): $0.36 per million
- **Egress**: Free (unlimited)

### Protecting Against Cost Overruns

1. **Bucket Privacy**: Private buckets prevent unauthorized access
2. **Token Scoping**: Bucket-specific tokens limit damage if compromised
3. **Monitoring**: Set up Cloudflare alerts for unusual activity
4. **Rate Limiting**: Consider Cloudflare Workers if you need additional protection

### Setting Up Alerts

1. Go to **Cloudflare Dashboard** → **Notifications**
2. Create alert for:
   - **R2 Operations** exceeding threshold
   - **R2 Storage** exceeding threshold
   - Unusual activity patterns

### Monitoring Usage

1. Go to **R2** → **Analytics**
2. Monitor:
   - Operations per day
   - Storage usage
   - Egress (should be minimal for state storage)

## Troubleshooting

### "Access Denied" Errors
- Verify R2 token has correct permissions
- Check bucket name matches exactly
- Ensure token is scoped to the correct bucket
- Verify Account ID in endpoint URL

### "Bucket Not Found" Errors
- Verify bucket name is correct
- Check bucket exists in your account
- Ensure you're using the correct Account ID

### High Costs
- Check R2 Analytics for unusual activity
- Verify bucket is private
- Review token permissions
- Consider rotating tokens if compromised

## Additional Security (Optional)

### Cloudflare Access (Enterprise Feature)
If you have Cloudflare Access, you can add an additional layer:
- Create Access policy for R2 bucket
- Require authentication for bucket access
- This adds extra protection but may not be necessary for private buckets

### State Encryption at Rest
R2 automatically encrypts data at rest. For additional security:
- Use Terraform's built-in encryption (if available)
- Consider encrypting state files before upload (advanced)

## Migration from Other Backends

If migrating from Terraform Cloud or S3:

1. Export current state: `terraform state pull > current-state.json`
2. Configure R2 backend in `terraform.tf`
3. Initialize: `terraform init -migrate-state`
4. Verify: `terraform state list`
5. Remove old backend configuration

## Backup Strategy

While R2 is durable, consider:
- Enabling bucket versioning (if available)
- Periodic state backups to another location
- State file snapshots before major changes

