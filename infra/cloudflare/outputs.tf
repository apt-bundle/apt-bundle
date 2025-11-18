output "root_domain_record" {
  description = "Root domain DNS record details"
  value = {
    name    = cloudflare_record.root.hostname
    type    = cloudflare_record.root.type
    content = cloudflare_record.root.content
    proxied = cloudflare_record.root.proxied
  }
}

output "www_record" {
  description = "WWW subdomain DNS record details"
  value = {
    name    = cloudflare_record.www.hostname
    type    = cloudflare_record.www.type
    content = cloudflare_record.www.content
    proxied = cloudflare_record.www.proxied
  }
}

output "all_records" {
  description = "Summary of all managed DNS records"
  value = {
    root = cloudflare_record.root.hostname
    www  = cloudflare_record.www.hostname
  }
}

