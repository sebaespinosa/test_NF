# Nginx Proxy Manager Integration Guide

## Overview

This guide covers the integration of **Nginx Proxy Manager (NPM)** into the Irrigation Analytics API stack. NPM provides a web-based interface for managing reverse proxy configurations, SSL certificates, and access control.

**Key Features:**
- Web UI for managing proxy hosts (no manual nginx config editing)
- Automatic Let's Encrypt SSL certificate provisioning and renewal
- HTTP→HTTPS redirects with force SSL
- Access control lists and basic authentication
- Stream forwarding for TCP/UDP services
- Custom SSL certificate support

**Architecture:**
- NPM runs as a Docker container
- API server runs on the **host machine** (not containerized)
- NPM proxies external requests to `host.docker.internal:8080` (the host-run API)
- Ports: 80 (HTTP), 443 (HTTPS), 81 (Admin UI)

---

## Prerequisites

- Docker and Docker Compose installed
- API server running on host at `http://localhost:8080`
- For local development: ability to edit `/etc/hosts`
- For production: DNS A record pointing to server IP

---

## Local Development Setup

### Step 1: Start Nginx Proxy Manager

NPM is included in the `docker-compose.yml` file. Start all services:

```bash
docker-compose up -d
```

Or start only NPM and its dependencies:

```bash
docker-compose up -d npm
```

Verify NPM is running:

```bash
docker-compose ps npm
```

### Step 2: Access NPM Admin UI

Open your browser and navigate to:

```
http://localhost:81
```

**⚠️ Important:** You will be prompted to create credentials on first login. Use a strong password.

### Step 3: Configure Local DNS

Add the following entry to your `/etc/hosts` file to resolve `irrigation.local` to your local machine:

```bash
# Edit hosts file (macOS/Linux)
sudo nano /etc/hosts

# Add this line:
127.0.0.1 irrigation.local
```

**Windows:** Edit `C:\Windows\System32\drivers\etc\hosts` with Administrator privileges.

Save and close the file. Verify resolution:

```bash
ping irrigation.local
# Should resolve to 127.0.0.1
```

### Step 4: Create Self Signed Certificate

See Appendix A

### Step 5: Create Proxy Host for API

In the NPM Admin UI:

1. **Navigate to:** Hosts → Certificates → Add Certificate → Custom Certificate

2. **Details Tab:**
   - **Name:** `irrigation.local SelfSigned`
   - **Certificate Key:** Paste contents or select file `irrigation.local.key`
   - **Certificate:** Paste contents or select file `irrigation.local.crt`
   - **Intermediate Certificate:** Leave blank

3. **Navigate to:** Hosts → Proxy Hosts → Add Proxy Host

4. **Details Tab:**
   - **Domain Names:** `irrigation.local`
   - **Scheme:** `http`
   - **Forward Hostname / IP:** `host.docker.internal`
   - **Forward Port:** `8080`
   - **Cache Assets:** ✅ (optional, improves performance)
   - **Block Common Exploits:** ✅ (recommended)
   - **Websockets Support:** ✅ (if API uses WebSockets)

5. **SSL Tab (Self-Signed for Local Dev):**
   - **SSL Certificate:** Select "irrigation.local SelfSigned"
   - **Force SSL:** ✅ (redirects HTTP → HTTPS)
   - **HTTP/2 Support:** ✅
   - **HSTS Enabled:** ✅ (optional, adds Strict-Transport-Security header)

6. Click **Save**

### Step 6: Test Local HTTPS Access

Open your browser and navigate to:

```
https://irrigation.local/health
```

**Expected Results:**
- ✅ HTTP requests to `http://irrigation.local` redirect to HTTPS
- ✅ HTTPS requests are proxied to `http://host.docker.internal:8080` (your host-run API)
- ⚠️ Browser will show "Not Secure" warning for self-signed certificate (expected for local dev)

**Accept the self-signed certificate** in your browser:
- Chrome/Edge: Click "Advanced" → "Proceed to irrigation.local (unsafe)"
- Firefox: Click "Advanced" → "Accept the Risk and Continue"

**Verify API Response:**

```bash
# Test HTTP→HTTPS redirect
curl -I http://irrigation.local
# Should return 301 or 302 redirect to https://irrigation.local

# Test HTTPS (ignore cert warning)
curl -k https://irrigation.local/health
# Should return: {"status":"healthy","message":"service is running","version":"..."}
```

---

## Production-Like Setup (with Let's Encrypt)

For production or testing with real SSL certificates, use Let's Encrypt.

### Prerequisites

- **Public domain name** (e.g., `api.yourdomain.com`)
- **DNS A record** pointing to your server's public IP
- **Ports 80 and 443** open in firewall (for Let's Encrypt validation and HTTPS traffic)

### Step 1: Verify DNS Resolution

Ensure your domain resolves to your server's public IP:

```bash
dig api.yourdomain.com +short
# Should return your server's public IP
```

### Step 2: Create Proxy Host with Let's Encrypt

In the NPM Admin UI:

1. **Navigate to:** Hosts → Proxy Hosts → Add Proxy Host

2. **Details Tab:**
   - **Domain Names:** `api.yourdomain.com`
   - **Scheme:** `http`
   - **Forward Hostname / IP:** `host.docker.internal`
   - **Forward Port:** `8080`
   - **Cache Assets:** ✅
   - **Block Common Exploits:** ✅
   - **Websockets Support:** ✅

3. **SSL Tab (Let's Encrypt):**
   - **SSL Certificate:** Select "Request a new SSL Certificate"
   - **Force SSL:** ✅ (mandatory for production)
   - **HTTP/2 Support:** ✅
   - **HSTS Enabled:** ✅
   - **HSTS Subdomains:** ✅ (if applicable)
   - **Email Address for Let's Encrypt:** `your-email@yourdomain.com`
   - **I Agree to the Let's Encrypt Terms of Service:** ✅

4. Click **Save**

NPM will:
- Request a certificate from Let's Encrypt via HTTP-01 challenge
- Automatically configure SSL for your domain
- Set up automatic renewal (certificates are valid for 90 days)

### Step 3: Verify Production SSL

```bash
# Test HTTP→HTTPS redirect
curl -I http://api.yourdomain.com
# Should return 301 redirect to https://api.yourdomain.com

# Test HTTPS with valid certificate
curl https://api.yourdomain.com/health
# Should return API response without certificate warnings

# Check SSL certificate
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com </dev/null | grep -A 2 "Verify return code"
# Should show: Verify return code: 0 (ok)
```

### Step 4: Monitor Certificate Renewal

NPM automatically renews Let's Encrypt certificates 30 days before expiration. Check renewal status:

- **NPM Admin UI:** SSL Certificates → View certificate details
- **Logs:** `docker-compose logs -f npm`

---

## Configuration Reference

### Docker Compose Service

The `npm` service in `docker-compose.yml`:

```yaml
npm:
  image: jc21/nginx-proxy-manager:latest
  container_name: npm
  ports:
    - "80:80"    # HTTP (public)
    - "443:443"  # HTTPS (public)
    - "81:81"    # Admin UI (restrict in production)
  environment:
    - DISABLE_IPV6=true
  volumes:
    - npm_data:/data
    - npm_letsencrypt:/etc/letsencrypt
  extra_hosts:
    - "host.docker.internal:host-gateway"
  restart: unless-stopped
  networks:
    - irrigation-network

volumes:
  npm_data:
  npm_letsencrypt:
```

**Key Configuration:**
- `extra_hosts`: Maps `host.docker.internal` to the host machine's gateway IP (allows NPM to reach the host-run API)
- `DISABLE_IPV6`: Prevents IPv6 binding issues on some systems
- Volumes:
  - `npm_data`: Stores NPM configuration, proxy host settings, and user data
  - `npm_letsencrypt`: Stores Let's Encrypt certificates and renewal metadata


### Security Considerations

**Production Deployments:**

1. **Restrict Admin UI Access:**
   - Change port `81:81` to `127.0.0.1:81:81` (local access only)
   - Use SSH tunneling for remote access: `ssh -L 8081:localhost:81 user@server`
   - Or set up a separate proxy host with basic authentication

2. **Enable HSTS:**
   - Force browsers to always use HTTPS
   - Prevents downgrade attacks

3. **Block Common Exploits:**
   - Enable in NPM to block known attack patterns

4. **Rate Limiting:**
   - Add custom nginx config to prevent abuse

5. **Firewall Rules:**
   - Allow ports 80/443 from public internet
   - Restrict port 81 to trusted IPs or localhost

### Common NPM Configurations

**1. CORS Headers (for API):**

In Proxy Host → Advanced tab:

```nginx
location / {
  add_header Access-Control-Allow-Origin *;
  add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
  add_header Access-Control-Allow-Headers "Authorization, Content-Type";
  proxy_pass http://host.docker.internal:8080;
}
```

**2. Rate Limiting:**

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location / {
  limit_req zone=api_limit burst=20 nodelay;
  proxy_pass http://host.docker.internal:8080;
}
```

**3. Custom Headers:**

```nginx
location / {
  proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_pass http://host.docker.internal:8080;
}
```

---

## Appendix A: Generate Self-Signed Certificate with OpenSSL

For local development, you can generate a self-signed certificate:

```bash
# Create directory for certificates
mkdir -p ./ssl

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
  -keyout ./ssl/irrigation.local.key \
  -out ./ssl/irrigation.local.crt \
  -subj "/CN=irrigation.local" \
  -addext "subjectAltName=DNS:irrigation.local,DNS:*.irrigation.local"
```

---

## References

- [Nginx Proxy Manager Documentation](https://nginxproxymanager.com/guide/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Nginx Configuration Reference](https://nginx.org/en/docs/)
- [Docker Extra Hosts](https://docs.docker.com/compose/compose-file/compose-file-v3/#extra_hosts)

---

## Support

For issues specific to NPM:
- NPM GitHub: https://github.com/NginxProxyManager/nginx-proxy-manager
- NPM Community Forum: https://github.com/NginxProxyManager/nginx-proxy-manager/discussions

For issues with the Irrigation Analytics API:
- See main [README.md](../README.md) and [IntegrationTesting.md](./IntegrationTesting.md)
