# R2 Backend Quick Start

This is a quick reference for setting up Cloudflare R2 as Terraform state backend. For detailed security information, see [r2-setup.md](r2-setup.md).

## Quick Setup Checklist

### 1. Create R2 Bucket

```bash
# Generate a secure, non-obvious bucket name
BUCKET_NAME=$(openssl rand -hex 8 | sed 's/^/tf-state-/')
echo "Bucket name: $BUCKET_NAME"
# Save this value - you'll need it for GitHub Secrets
```

1. Go to Cloudflare Dashboard → **R2** → **Create bucket**
2. Use the generated bucket name (e.g., `tf-state-a7f3b2c9d1e4f5g6`)
3. **Keep public access DISABLED** (default)

### 2. Create R2 API Token

1. Go to **R2** → **Manage R2 API Tokens** → **Create API token**
2. Configure:
   - **Token name**: `terraform-state-r2`
   - **Permissions**: ✅ Object Read, ✅ Object Write
   - **Bucket restriction**: Select your specific bucket
3. **Copy both values immediately**:
   - Access Key ID
   - Secret Access Key

### 3. Get Account ID

1. In Cloudflare Dashboard, select any domain
2. Right sidebar → **API** section → Copy **Account ID**
3. Or: R2 → Your bucket → Settings → Account ID

### 4. Add GitHub Secrets

Go to: https://github.com/apt-bundle/apt-bundle/settings/secrets/actions

Add these secrets:

| Secret Name | Value |
|------------|-------|
| `R2_ACCESS_KEY_ID` | Your R2 Access Key ID |
| `R2_SECRET_ACCESS_KEY` | Your R2 Secret Access Key |
| `R2_BUCKET_NAME` | Your bucket name (e.g., `tf-state-a7f3b2c9d1e4f5g6`) |
| `R2_ENDPOINT` | `https://<your-account-id>.r2.cloudflarestorage.com` |

### 5. Test Locally (Optional)

```bash
cd infra/cloudflare

# Create backend config
cat > backend.hcl <<EOF
bucket                      = "your-bucket-name"
endpoint                    = "https://your-account-id.r2.cloudflarestorage.com"
access_key                  = "your-access-key-id"
secret_key                  = "your-secret-access-key"
EOF

# Initialize Terraform
terraform init -backend-config=backend.hcl

# Verify it works
terraform plan -var="cloudflare_zone_id=your-zone-id"
```

**Important**: `backend.hcl` is gitignored - never commit it!

### 6. Verify GitHub Actions

After adding secrets, push a change to trigger the workflow:

```bash
git add infra/cloudflare/
git commit -m "Configure R2 backend"
git push
```

The GitHub Actions workflows will automatically:
- Generate `backend.hcl` from secrets
- Initialize Terraform with R2 backend
- Run plan/apply as configured

## Security Reminders

✅ **DO:**
- Use non-obvious bucket names
- Scope tokens to specific buckets
- Keep buckets private
- Monitor R2 usage

❌ **DON'T:**
- Use descriptive bucket names
- Create account-wide tokens
- Commit `backend.hcl` or credentials
- Enable public access

## Troubleshooting

**"Access Denied"**: Check token permissions and bucket name  
**"Bucket Not Found"**: Verify bucket name and Account ID  
**"Backend initialization failed"**: Check all GitHub Secrets are set correctly

For detailed troubleshooting, see [r2-setup.md](r2-setup.md).

