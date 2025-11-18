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

# Note: APT repository is served from apt-bundle.org/repo/ path instead of
# repo.apt-bundle.org subdomain due to GitHub Pages limitation of one custom
# domain per repository. The repo subdomain DNS record has been removed.

