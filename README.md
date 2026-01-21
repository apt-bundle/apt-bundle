# apt-bundle

> A **simple, declarative tool** to manage APT package sets and repositories on Debian-based systems — ideal for CI, repeatable environment setup, and lightweight provisioning.
> *Not a full configuration management system.*

## 📌 What This Is

`apt-bundle` helps you define your system packages, PPAs, custom repos, and GPG keys in a **single text file (`Aptfile`)**, and install them with one command.
It’s designed to reduce boilerplate around Debian/Ubuntu package installs — especially in CI runners, ephemeral VMs, Docker build contexts, or simple developer onboarding scripts.

> ⚠️ This is **not a replacement for configuration management tools** like Ansible, nor a substitute for language-specific package managers. It focuses on a *declarative list of APT packages*. (You decide where full config management is a better fit.)

## 🎯 What apt-bundle *Does* Well

* **Declarative system dependencies:** Put your packages in an `Aptfile` for readability and sharing.
* **Idempotent installs:** You can run it repeatedly without reshuffling your system.
* **Repository & PPA support:** Add Ubuntu PPAs and custom APT repos.
* **Simple CLI:** Minimal surface area to learn and automate.

## 🧠 What apt-bundle *Doesn’t* Aim To Be

Use cases that *don’t align* with this tool:

* Full **configuration management** (services, users, networking).
* Complex system state enforcement across fleets.
* Docker image optimization (it helps readability, but doesn’t reduce layer size).
* Replacement for language-specific managers (e.g., `pip`, `asdf`, `nvm`).

If you need those, tools like **Ansible**, **Nix**, or language version managers are more appropriate.

---

## 🚀 Quick Start

### 1. Install apt-bundle

#### Recommended: Quick Script

```bash
curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | sudo bash
```

(This installs the latest Go-built `.deb` for your architecture.)

#### Alternate: From `.deb` Release

Download the latest release `.deb` and install with `dpkg -i`, then resolve deps with `apt install -f`.

#### From Source

```bash
git clone https://github.com/apt-bundle/apt-bundle.git
cd apt-bundle
make build
sudo make install
```

Binary ends up in `/usr/local/bin/apt-bundle`.


## 📦 Common Commands

```bash
# Install packages from Aptfile (default: ./Aptfile)
sudo apt-bundle

# Explicit install command
sudo apt-bundle install

# Use a custom file
sudo apt-bundle --file /path/to/Aptfile

# Skip updating package lists (CI optimization)
sudo apt-bundle --no-update

# Check installed packages (non-root)
apt-bundle check

# Export current system packages to Aptfile
apt-bundle dump > Aptfile
```

## 🧩 Typical Use Cases

* **CI Runners:** Replace long, error-prone `apt-get install …` steps with a maintained list.
* **Developer Onboarding:** Keep a canonical set of system dependencies in your repo.
* **Local Bootstrap / Fresh VMs:** One command to align machines.
* **Docker Build Clarity:** Improves readability, though classic Dockerfile patterns still control image layering.

---

## 📄 Aptfile Format

Create an `Aptfile` describing your desired packages.

**Example**

```
# Core tools
apt git
apt curl
apt build-essential

# Specific version
apt "nano=2.9.3-2"

# Add PPA and packages from it
ppa ppa:ondrej/php
apt php8.1 php8.1-cli php8.1-fpm

# Add custom repository
key https://download.docker.com/linux/ubuntu/gpg
deb "[arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
apt docker-ce docker-ce-cli containerd.io
```

(Here `ppa` adds a Personal Package Archive; the system will fetch from that source.)

---

## 🧭 Summary

`apt-bundle` is:

✅ **Declarative** — keep system package lists in one place.
✅ **Repeatable** — useful in CI, bootstrap, and simple reproducible environments.
⚠️ **Not a full config management stack** — know its boundaries.

---

## 📜 License

This project is licensed under **Apache-2.0**.
