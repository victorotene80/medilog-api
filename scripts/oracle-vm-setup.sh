#!/usr/bin/env bash
# One-time bootstrap for a fresh Oracle Cloud Always Free Ampere A1 VM
# (Ubuntu 22.04/24.04 aarch64 image assumed). Run once, as the ubuntu user,
# right after first SSH login:
#
#   ssh ubuntu@<VM_PUBLIC_IP>
#   curl -fsSL https://raw.githubusercontent.com/<you>/medilog-api/main/scripts/oracle-vm-setup.sh | bash
#   # or: scp the repo up and run it locally
#
# Installs Docker, opens the host firewall (Oracle images ship with iptables
# rules that DROP everything not explicitly allowed, on top of the OCI
# Security List / NSG — both layers need the same ports opened), and clones
# the repo.
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/CHANGE_ME/medilog-api.git}"
REPO_DIR="$HOME/medilog-api"

echo "=== Installing Docker ==="
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker "$USER"
  echo ">>> Log out and back in for the docker group membership to take effect."
fi

echo "=== Opening host firewall (iptables) for HTTP/HTTPS/SSH ==="
# Oracle's default Ubuntu image firewalls the host itself in addition to the
# OCI-level Security List — both must allow 80/443 or Caddy is unreachable
# from the internet even though the container is listening fine.
sudo iptables -I INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 443 -j ACCEPT
sudo netfilter-persistent save 2>/dev/null || sudo iptables-save | sudo tee /etc/iptables/rules.v4 >/dev/null 2>&1 || true

echo ">>> Also open ingress rules for TCP 80 and 443 in the OCI console:"
echo ">>>   VCN -> your subnet -> Security List (or the VM's NSG) -> Add Ingress Rule"
echo ">>>   Source CIDR 0.0.0.0/0, Destination Port 80 and 443"

echo "=== Cloning repo ==="
if [ ! -d "$REPO_DIR" ]; then
  git clone "$REPO_URL" "$REPO_DIR"
fi

echo "=== Done ==="
echo "Next steps:"
echo "  1. cd $REPO_DIR"
echo "  2. cp .env.prod.example .env && chmod 600 .env && edit it"
echo "  3. Point your DuckDNS domain at this VM's public/reserved IP"
echo "  4. ./scripts/deploy.sh"
