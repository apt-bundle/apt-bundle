# Technical Specification: apt-bundle

## Implementation Overview

apt-bundle uses modern APT conventions:

- **Custom repositories:** Stored as DEB822-format `.sources` files in `/etc/apt/sources.list.d/` (not legacy `.list` one-line format).
- **GPG keys:** Stored in `/etc/apt/keyrings/` (not `/etc/apt/trusted.gpg.d/`). Keys are scoped per-repository via the `Signed-By` field in DEB822 stanzas, rather than globally trusted.
- **Key URLs:** Only `https://` URLs are accepted; `http://` and `file://` are rejected for security.

---

## Part 1: Aptfile Specification

### Overview
The Aptfile is a line-oriented, UTF-8 encoded text file. The default filename is Aptfile.
Comments: Lines beginning with # are ignored.
Empty Lines: Empty lines are ignored.
Directives: Each line consists of a directive (keyword) followed by a quoted or unquoted argument. Quotes (" or ') are required if the argument contains spaces or special characters.

### Directives
`apt` Specifies a standard apt package to be installed.
Syntax: `apt <package-name>[=<version-string>]`
Logic: Corresponds to `apt-get install -y <package-name>[=<version-string>]`.
Examples:

```
# Installs the latest version

apt vim
apt curl
apt build-essential

# Installs a specific version
apt "nano=2.9.3-2"
apt "docker-ce=5:19.03.13~3-0~ubuntu-focal"
```

`ppa` Specifies a Personal Package Archive (PPA) to be added.
Syntax: ppa <ppa:user/repo>
Logic: Corresponds to add-apt-repository -y <ppa:user/repo>.
Note: This directive implicitly handles adding the PPA's GPG key.
Example:

```
ppa ppa:ondrej/php
ppa ppa:git-core/ppa
```

`deb` Specifies a full deb repository line to be added to /etc/apt/sources.list.d/.
Syntax: deb "<full-repository-line>"
Logic: The tool creates a DEB822-format .sources file (e.g., apt-bundle-<hash>.sources) in /etc/apt/sources.list.d/.
Note: This directive only adds the repository line. It does not add the GPG key. The key directive must be used separately. When a key directive precedes a deb directive, the key path is used for the Signed-By field in the DEB822 stanza.
Example:

```
# Google Chrome
deb "[arch=amd64] [http://dl.google.com/linux/chrome/deb/](http://dl.google.com/linux/chrome/deb/) stable main"

# Docker
deb "[arch=amd64] [https://download.docker.com/linux/ubuntu](https://download.docker.com/linux/ubuntu) focal stable"
```


`key` Specifies a GPG key URL to be downloaded and added to the keyring.
Syntax: key <url-to-key>
Logic: The tool downloads the key, dearmors it if needed, and saves it to /etc/apt/keyrings/ (e.g., apt-bundle-<hash>.gpg). The key path is used for the Signed-By field when adding deb repositories, scoping trust to that repository rather than globally.
Example:

```
# Google Chrome Key
key [https://dl.google.com/linux/linux_signing_key.pub](https://dl.google.com/linux/linux_signing_key.pub)

# Docker Key
key [https://download.docker.com/linux/ubuntu/gpg](https://download.docker.com/linux/ubuntu/gpg)
```


### Example Aptfile

```
# This is a sample Aptfile

# Core dev tools
apt build-essential
apt curl
apt git
apt vim
apt htop

# Specific version of nano
apt "nano=2.9.3-2"

# PHP from a PPA
ppa ppa:ondrej/php
apt php8.1
apt php8.1-cli

# Install Docker
key [https://download.docker.com/linux/ubuntu/gpg](https://download.docker.com/linux/ubuntu/gpg)
deb "[arch=amd64] [https://download.docker.com/linux/ubuntu](https://download.docker.com/linux/ubuntu) focal stable"
apt docker-ce
apt docker-ce-cli
apt containerd.io
```
