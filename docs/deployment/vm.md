← [Back to Index](../INDEX.md)

# Linux VM Deployment Guide (systemd)

> Production-oriented deployment of the OathMesh issuer on a Linux VM using `systemd`.

## 1) Host prerequisites

- Linux VM with `systemd` (Ubuntu/Debian/RHEL-family)
- Dedicated service user (example: `oathmesh`)
- TLS termination at LB/reverse proxy (issuer URL should be `https://...` in production)
- Redis available for replay cache/stateful revocation (`REDIS_URL`)

Recommended layout:

```text
/opt/oathmesh/bin/oathmesh
/etc/oathmesh/issuer.env
/etc/oathmesh/private.pem
```

Build and install binary (binary name matches this repo):

```bash
make build
sudo install -d -m 0755 /opt/oathmesh/bin
sudo install -m 0755 ./bin/oathmesh /opt/oathmesh/bin/oathmesh
```

## 2) Environment file and secret handling

Create runtime environment file (never commit real values):

```bash
sudo install -d -m 0750 /etc/oathmesh
sudo cp scripts/systemd/oathmesh-issuer.env.example /etc/oathmesh/issuer.env
sudo chown root:oathmesh /etc/oathmesh/issuer.env
sudo chmod 0640 /etc/oathmesh/issuer.env
```

Generate and restrict key material:

```bash
openssl genpkey -algorithm Ed25519 -out private.pem
sudo mv private.pem /etc/oathmesh/private.pem
sudo chown root:oathmesh /etc/oathmesh/private.pem
sudo chmod 0640 /etc/oathmesh/private.pem
```

Hard requirements:

- Never commit `.env`/`issuer.env`, private keys, or mint secrets.
- Prefer secret manager injection over static files where possible.
- Keep `/etc/oathmesh` readable only by root + service group.

## 3) Install and run the systemd service

Copy the example unit:

```bash
sudo cp scripts/systemd/oathmesh-issuer.service /etc/systemd/system/oathmesh-issuer.service
sudo systemctl daemon-reload
sudo systemctl enable --now oathmesh-issuer
```

Check status:

```bash
systemctl status oathmesh-issuer --no-pager
journalctl -u oathmesh-issuer -n 100 --no-pager
```

## 4) Health checks and runtime verification

Run local health checks:

```bash
curl -fsS http://127.0.0.1:4000/healthz
curl -fsS http://127.0.0.1:4000/.well-known/jwks.json | head -c 200
```

Optional metrics check:

```bash
curl -fsS http://127.0.0.1:4000/metrics | head -n 20
```

## 5) Log management and rotation

The example unit logs to journald. Rotate and cap journal usage:

```bash
sudo install -d -m 0755 /etc/systemd/journald.conf.d
cat <<'EOF' | sudo tee /etc/systemd/journald.conf.d/oathmesh.conf >/dev/null
[Journal]
SystemMaxUse=1G
RuntimeMaxUse=256M
MaxRetentionSec=14day
EOF
sudo systemctl restart systemd-journald
```

If you choose file audit logs (`OATHMESH_AUDIT_SINK=file`), add logrotate:

```conf
/var/log/oathmesh/audit.ndjson {
  daily
  rotate 14
  compress
  missingok
  notifempty
  create 0640 root oathmesh
}
```

## 6) Firewall and network hardening

- Only expose issuer port to trusted networks/LB.
- Restrict egress to required dependencies (Redis, DNS, NTP, telemetry).
- Keep SSH locked down (key auth, allowlist, no password login).

Example (`ufw`):

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from 10.0.0.0/8 to any port 4000 proto tcp
sudo ufw allow OpenSSH
sudo ufw enable
```

## 7) Backup and restore

Back up at minimum:

- `/etc/oathmesh/issuer.env` (securely, encrypted)
- `/etc/oathmesh/private.pem` (encrypted, access-controlled)
- Redis persistence data used for revocation/replay continuity

Restore checklist:

1. Restore key + env files with original permissions.
2. Restore Redis data/snapshot.
3. Start issuer: `sudo systemctl restart oathmesh-issuer`.
4. Validate `/healthz` and JWKS endpoint.
5. Mint and verify a test token end-to-end.

## 8) Upgrade procedure (rolling-forward)

1. Build new binary: `make build`
2. Install atomically:
   ```bash
   sudo install -m 0755 ./bin/oathmesh /opt/oathmesh/bin/oathmesh.new
   sudo mv /opt/oathmesh/bin/oathmesh.new /opt/oathmesh/bin/oathmesh
   ```
3. Validate unit/env config:
   ```bash
   sudo systemd-analyze verify /etc/systemd/system/oathmesh-issuer.service
   ```
4. Restart service:
   ```bash
   sudo systemctl restart oathmesh-issuer
   ```
5. Post-upgrade checks:
   ```bash
   systemctl is-active oathmesh-issuer
   curl -fsS http://127.0.0.1:4000/healthz
   ```
6. If failed, roll back with previous binary and restart.

## Related

- [Docker Compose Deployment](docker-compose.md)
- [Kubernetes Deployment Guide](kubernetes.md)
- [TLS Configuration Guide](tls.md)
- [Issuer Configuration Reference](../config/issuer-config.md)
