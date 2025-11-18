# DNS Records for apt-bundle.org

# Root domain CNAME to GitHub Pages
# This record is proxied through Cloudflare for DDoS protection and CDN
resource "cloudflare_record" "root" {
  zone_id = var.cloudflare_zone_id
  name    = "@"
  type    = "CNAME"
  content = var.github_pages_target
  proxied = true
  comment = "Root domain pointing to GitHub Pages for main website"
  ttl     = 1 # Auto TTL when proxied
}

# www subdomain CNAME to GitHub Pages
# This record is proxied through Cloudflare for DDoS protection and CDN
resource "cloudflare_record" "www" {
  zone_id = var.cloudflare_zone_id
  name    = "www"
  type    = "CNAME"
  content = var.github_pages_target
  proxied = true
  comment = "www subdomain pointing to GitHub Pages for main website"
  ttl     = 1 # Auto TTL when proxied
}

# APT repository subdomain CNAME to GitHub Pages
# IMPORTANT: This record is NOT proxied (proxied = false)
# Reason: APT package managers need direct DNS resolution and don't work
# properly with Cloudflare's proxy. APT clients expect direct access to
# the origin server and may fail with SSL/TLS issues when proxied.
resource "cloudflare_record" "repo" {
  zone_id = var.cloudflare_zone_id
  name    = "repo"
  type    = "CNAME"
  content = var.github_pages_target
  proxied = false
  comment = "APT repository subdomain - NOT proxied for APT client compatibility"
  ttl     = 300 # 5 minutes - standard TTL for non-proxied records
}

