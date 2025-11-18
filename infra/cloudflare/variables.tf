variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for apt-bundle.org domain"
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^[a-f0-9]{32}$", var.cloudflare_zone_id))
    error_message = "Zone ID must be a 32-character hexadecimal string."
  }
}

variable "domain" {
  description = "Root domain name"
  type        = string
  default     = "apt-bundle.org"

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*\\.[a-z]{2,}$", var.domain))
    error_message = "Domain must be a valid domain name."
  }
}

variable "github_pages_target" {
  description = "GitHub Pages CNAME target"
  type        = string
  default     = "apt-bundle.github.io"
}

