## 🎮 Discord Activity Setup - Kubex Studio

### Overview

Integre todo o ecossistema Kubex (GoBE, Grompt, Analyzer, GemX) como um **Discord Activity** embedado diretamente no servidor Discord.

```
┌─────────────────────────────────────────────────┐
│         Discord Server - Kubex HQ               │
├─────────────────────────────────────────────────┤
│  🎮 Activity: "Kubex Studio"                    │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │  🔐 OAuth 2.1 Login                      │  │
│  │  (Automatic via Discord)                 │  │
│  └──────────────────────────────────────────┘  │
│                    ▼                             │
│  ┌──────────────────────────────────────────┐  │
│  │  Tabs: [GoBE] [Grompt] [Analyzer]       │  │
│  │                                          │  │
│  │  /web/dashboard  → GoBE Metrics         │  │
│  │  /web/grompt     → Prompt Studio        │  │
│  │  /web/analyzer   → Code Analysis        │  │
│  │  /web/gemx       → Image Generation     │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

1. **Discord Developer Account**
   - Go to https://discord.com/developers/applications
   - Create new Application

2. **GoBE Server Running**
   - Public URL (e.g., `https://gobe.kubex.io`)
   - SSL/TLS certificate configured
   - OAuth 2.1 enabled

3. **Ecosystem Services**
   - Grompt running on port 8080
   - Analyzer running on port 8081
   - GemX running on port 8082 (optional)

---

## 🚀 Step-by-Step Setup

### Step 1: Create Discord Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"**
3. Name it: `Kubex Studio`
4. Save Application ID (you'll need this)

### Step 2: Configure OAuth2

1. In your Discord app, go to **OAuth2** tab
2. Add Redirect URLs:
   ```
   https://gobe.kubex.io/auth/discord/callback
   https://gobe.kubex.io/web
   ```

3. Select Scopes:
   - ✅ `identify` - Get user ID
   - ✅ `email` - Get user email
   - ✅ `guilds` - Get user's servers

4. Copy **Client ID** and **Client Secret**

### Step 3: Create Activity

1. Go to **Activities** tab in Discord app
2. Click **"New Activity"**
3. Configure:
   ```json
   {
     "name": "Kubex Studio",
     "description": "Access GoBE, Grompt, Analyzer, and GemX",
     "type": "EMBEDDED",
     "default_orientation": "LANDSCAPE",
     "supported_platforms": ["DESKTOP", "MOBILE"]
   }
   ```

4. Set **Activity URL Mappings**:
   ```
   / → https://gobe.kubex.io/web
   ```

5. Upload Activity Icon (512x512 PNG)

### Step 4: Configure GoBE

Create/update `config/discord_activity.yaml`:

```yaml
discord:
  activity:
    enabled: true
    client_id: "YOUR_DISCORD_CLIENT_ID"
    client_secret: "YOUR_DISCORD_CLIENT_SECRET"
    redirect_uri: "https://gobe.kubex.io/auth/discord/callback"
    scopes:
      - identify
      - email
      - guilds

oauth:
  enabled: true
  jwt_secret: "your-super-secret-jwt-key-change-in-production"
  token_expiry: 86400  # 24 hours

proxy:
  grompt_url: "http://localhost:8080"
  analyzer_url: "http://localhost:8081"
  gemx_url: "http://localhost:8082"

web:
  require_auth: true
  allow_discord_iframe: true
  cors_origins:
    - "https://discord.com"
    - "https://ptb.discord.com"
    - "https://canary.discord.com"
```

### Step 5: Environment Variables

```bash
# Discord Activity
export DISCORD_CLIENT_ID="your_client_id_here"
export DISCORD_CLIENT_SECRET="your_client_secret_here"

# OAuth 2.1
export GOBE_JWT_SECRET="your-super-secret-jwt-key"
export GOBE_OAUTH_ENABLED=true

# Proxy Configuration
export GROMPT_URL=http://localhost:8080
export ANALYZER_URL=http://localhost:8081
export GEMX_URL=http://localhost:8082

# Server
export GOBE_PUBLIC_URL=https://gobe.kubex.io
export PORT=3666
```

### Step 6: Update GoBE Router

The routes are already configured in `/internal/app/router/web/web_routes.go`:

```go
// Routes are automatically configured:
// GET  /web              → Dashboard (OAuth protected)
// GET  /web/grompt/*     → Grompt proxy (OAuth protected)
// GET  /web/analyzer/*   → Analyzer proxy (OAuth protected)
// GET  /web/gemx/*       → GemX proxy (OAuth protected)

// OAuth routes:
// GET  /oauth/authorize  → OAuth 2.1 authorization
// POST /oauth/token      → Token exchange
```

### Step 7: Test Locally (Development)

1. **Start services:**
   ```bash
   # Terminal 1: Grompt
   cd /projects/kubex/grompt && ./grompt start -p 8080

   # Terminal 2: GoBE
   cd /projects/kubex/gobe && ./gobe start -p 3666
   ```

2. **Test OAuth flow:**
   ```bash
   # Generate auth code
   curl "http://localhost:3666/oauth/authorize?\
   client_id=YOUR_CLIENT_ID&\
   redirect_uri=http://localhost:3666/auth/callback&\
   code_challenge=CHALLENGE&\
   code_challenge_method=S256&\
   scope=read"
   ```

3. **Access web UI:**
   ```bash
   open http://localhost:3666/web
   ```

### Step 8: Deploy to Production

1. **Setup SSL/TLS:**
   ```bash
   # Using Let's Encrypt
   certbot certonly --standalone -d gobe.kubex.io
   ```

2. **Configure reverse proxy (Nginx):**
   ```nginx
   server {
       listen 443 ssl http2;
       server_name gobe.kubex.io;

       ssl_certificate /etc/letsencrypt/live/gobe.kubex.io/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/gobe.kubex.io/privkey.pem;

       # Headers for Discord iframe
       add_header X-Frame-Options "ALLOW-FROM https://discord.com";
       add_header Content-Security-Policy "frame-ancestors 'self' https://discord.com https://*.discord.com";

       location / {
           proxy_pass http://localhost:3666;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection 'upgrade';
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
           proxy_cache_bypass $http_upgrade;
       }
   }
   ```

3. **Start GoBE with production config:**
   ```bash
   export GOBE_ENV=production
   export GOBE_PUBLIC_URL=https://gobe.kubex.io
   ./gobe start -p 3666
   ```

### Step 9: Enable Activity in Discord

1. Go back to Discord Developer Portal
2. Your application → **Activities** tab
3. Click **"Publish Activity"**
4. Wait for approval (usually 24-48 hours)

### Step 10: Add to Discord Server

1. In Discord server, click **"+"** next to voice channel
2. Select **"Activities"**
3. Find **"Kubex Studio"**
4. Click to launch!

---

## 🎯 User Flow

```
User clicks "Kubex Studio" Activity in Discord
    ↓
Discord opens iframe: https://gobe.kubex.io/web
    ↓
GoBE checks OAuth token (from Discord)
    ↓
If not authenticated:
    ├─→ Redirect to /oauth/authorize
    ├─→ Discord OAuth login (automatic)
    └─→ Generate JWT token
    ↓
If authenticated:
    ├─→ Show dashboard with tabs
    ├─→ User clicks "Grompt" tab
    ├─→ GoBE proxies to localhost:8080
    └─→ Grompt UI loads inside Discord!
```

---

## 🔧 Troubleshooting

### Issue: "Refused to frame... X-Frame-Options"

**Solution:** Ensure CORS headers are set:
```go
// In discord_cors.go
c.Writer.Header().Set("X-Frame-Options", "ALLOW-FROM https://discord.com")
c.Writer.Header().Set("Content-Security-Policy",
    "frame-ancestors 'self' https://discord.com https://*.discord.com")
```

### Issue: "OAuth authorization failed"

**Solution:** Check redirect URI matches exactly:
```bash
# In Discord app OAuth settings
https://gobe.kubex.io/auth/discord/callback

# In config
redirect_uri: "https://gobe.kubex.io/auth/discord/callback"
```

### Issue: "Proxy connection refused"

**Solution:** Verify services are running:
```bash
# Check Grompt
curl http://localhost:8080/api/health

# Check logs
tail -f /var/log/gobe/proxy.log
```

### Issue: "CORS error in browser console"

**Solution:** Add Discord origin to CORS middleware:
```go
allowedOrigins := []string{
    "https://discord.com",
    "https://ptb.discord.com",
    "https://canary.discord.com",
}
```

---

## 📊 Monitoring

### Check OAuth tokens:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://gobe.kubex.io/auth/me
```

### Check proxy status:
```bash
curl https://gobe.kubex.io/web/grompt/api/health
curl https://gobe.kubex.io/web/analyzer/api/health
```

### View logs:
```bash
# GoBE logs
tail -f ~/.kubex/gobe/logs/server.log

# OAuth logs
tail -f ~/.kubex/gobe/logs/oauth.log

# Proxy logs
tail -f ~/.kubex/gobe/logs/proxy.log
```

---

## 🔒 Security Best Practices

1. **Always use HTTPS** in production
2. **Rotate JWT secrets** regularly
3. **Validate Discord origin** on all requests
4. **Rate limit** OAuth endpoints
5. **Log all authentication attempts**
6. **Use short-lived tokens** (24h max)
7. **Implement refresh tokens** for better UX

---

## 📚 API Reference

### OAuth Endpoints

#### GET /oauth/authorize
Initiates OAuth 2.1 PKCE flow.

**Parameters:**
- `client_id` (required): Discord application ID
- `redirect_uri` (required): Callback URL
- `code_challenge` (required): PKCE challenge
- `code_challenge_method`: `S256` or `plain`
- `scope`: Requested scopes
- `state`: CSRF token

**Response:**
```
302 Redirect to: {redirect_uri}?code={auth_code}&state={state}
```

#### POST /oauth/token
Exchanges authorization code for access token.

**Body:**
```json
{
  "grant_type": "authorization_code",
  "code": "AUTH_CODE",
  "code_verifier": "VERIFIER",
  "client_id": "CLIENT_ID",
  "redirect_uri": "CALLBACK_URL"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "scope": "identify email"
}
```

### Web Proxy Endpoints

All require OAuth Bearer token in `Authorization` header.

#### GET /web
Main dashboard

#### GET /web/grompt/*
Proxies to Grompt service

#### GET /web/analyzer/*
Proxies to Analyzer service

#### GET /web/gemx/*
Proxies to GemX service

---

## 🎨 Customization

### Custom Login Page

Edit `/internal/app/middlewares/web_auth.go`:
```go
func (m *WebAuthMiddleware) serveLoginPage(c *gin.Context) {
    // Customize HTML here
    html := `...your custom login page...`
    c.String(200, html)
}
```

### Custom Error Pages

Edit `/internal/app/router/proxy/web_proxy.go`:
```go
func createErrorHandler(serviceName string) {
    // Customize error pages here
}
```

---

## 📖 Related Documentation

- [OAuth 2.1 Specification](https://oauth.net/2.1/)
- [Discord Activities](https://discord.com/developers/docs/activities/overview)
- [GoBE CLAUDE.md](../CLAUDE.md)
- [MCP Integration Guide](./MCP_INTEGRATION.md)

---

**Version:** 1.3.5
**Last Updated:** 2025-01-20
**License:** MIT
