// Package middlewares provides authentication middleware for web UI
package middlewares

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// WebAuthConfig holds configuration for web authentication
type WebAuthConfig struct {
	JWTSecret     string
	SessionSecret string
	RequireAuth   bool // If false, auth is optional
	AllowDiscord  bool // Allow Discord OAuth tokens
	CertService   sci.ICertService
}

// WebAuthMiddleware handles authentication for web UI routes
type WebAuthMiddleware struct {
	config      WebAuthConfig
	certService sci.ICertService
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

// NewWebAuthMiddleware creates a new web auth middleware
func NewWebAuthMiddleware(config WebAuthConfig) *WebAuthMiddleware {
	middleware := &WebAuthMiddleware{}

	certService := config.CertService
	if certService == nil {
		defaultKeyPath := os.ExpandEnv(gl.DefaultGoBEKeyPath)
		defaultCertPath := os.ExpandEnv(gl.DefaultGoBECertPath)
		certService = crt.NewCertService(defaultKeyPath, defaultCertPath)
	}

	middleware.certService = certService

	if certService != nil {
		privKey, err := certService.GetPrivateKey() // pragma: allowlist secret
		if err != nil {
			gl.Log("error", "Failed to load RSA private key for web auth", "error", err)
		} else {
			middleware.privateKey = privKey // pragma: allowlist secret
		}

		pubKey, err := certService.GetPublicKey() // pragma: allowlist secret
		if err != nil {
			gl.Log("error", "Failed to load RSA public key for web auth", "error", err)
		} else {
			middleware.publicKey = pubKey
		}
	} else {
		gl.Log("warn", "Certificate service not configured for web auth; falling back to HMAC secret")
	}

	if middleware.privateKey != nil && middleware.publicKey != nil { // pragma: allowlist secret
		gl.Log("info", "Web auth middleware configured to validate JWT using RSA keys")
	} else {
		if config.JWTSecret == "" {
			keyringSecret, err := crt.GetOrGenPasswordKeyringPass("jwt_secret") // pragma: allowlist secret
			if err != nil {
				config.JWTSecret = "gobe-default-secret-change-in-production" // pragma: allowlist secret
				gl.Log("warn", "Using default JWT secret fallback - RSA keys unavailable", "error", err)
			} else {
				config.JWTSecret = keyringSecret                                                     // pragma: allowlist secret
				gl.Log("info", "Web auth middleware using keyring-managed secret for HMAC fallback") // pragma: allowlist secret
			}
		}

		if config.JWTSecret == "gobe-default-secret-change-in-production" { // pragma: allowlist secret
			gl.Log("warn", "Default JWT secret is active; configure RSA certificates or custom secret for production")
		}
	}

	config.CertService = certService
	middleware.config = config
	return middleware
}

// RequireAuth is the middleware function
func (m *WebAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if auth is required
		if !m.config.RequireAuth {
			c.Next()
			return
		}

		// Try to get token from multiple sources
		token := m.extractToken(c)

		if token == "" {
			m.serveLoginPage(c)
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.validateToken(token)
		if err != nil {
			gl.Log("warn", "Invalid auth token", "error", err, "ip", c.ClientIP())
			m.serveLoginPage(c)
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("authenticated", true)

		gl.Log("debug", "User authenticated", "user", claims["username"])

		c.Next()
	}
}

// OptionalAuth provides optional authentication
func (m *WebAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)

		if token != "" {
			claims, err := m.validateToken(token)
			if err == nil {
				c.Set("user_id", claims["user_id"])
				c.Set("username", claims["username"])
				c.Set("authenticated", true)
			}
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// extractToken extracts JWT token from request
func (m *WebAuthMiddleware) extractToken(c *gin.Context) string {
	// 1. Check Authorization header (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 2. Check cookie
	cookie, err := c.Cookie("gobe_token")
	if err == nil && cookie != "" {
		return cookie
	}

	// 3. Check query parameter (for Discord iframe)
	if token := c.Query("token"); token != "" {
		return token
	}

	// 4. Check X-Auth-Token header (custom)
	if token := c.GetHeader("X-Auth-Token"); token != "" {
		return token
	}

	return ""
}

// validateToken validates JWT token and returns claims
func (m *WebAuthMiddleware) validateToken(tokenString string) (jwt.MapClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	if m.publicKey != nil {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.publicKey, nil
		})
	} else {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.config.JWTSecret), nil
		})
	}

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token expired")
			}
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// serveLoginPage serves the login page for unauthenticated users
func (m *WebAuthMiddleware) serveLoginPage(c *gin.Context) {
	// Check if it's an API request (JSON)
	if c.GetHeader("Accept") == "application/json" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "Unauthorized",
			"message":   "Authentication required",
			"login_url": "/auth/login",
		})
		return
	}

	// Serve HTML login page
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoBE - Login Required</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            color: #333;
        }

        .login-container {
            background: white;
            padding: 3rem;
            border-radius: 1.5rem;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 400px;
            width: 90%;
            text-align: center;
        }

        .logo {
            font-size: 3rem;
            margin-bottom: 1rem;
        }

        h1 {
            color: #667eea;
            margin-bottom: 0.5rem;
            font-size: 2rem;
        }

        p {
            color: #666;
            margin-bottom: 2rem;
        }

        .login-form {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        input {
            padding: 1rem;
            border: 2px solid #e0e0e0;
            border-radius: 0.5rem;
            font-size: 1rem;
            transition: border-color 0.3s;
        }

        input:focus {
            outline: none;
            border-color: #667eea;
        }

        button {
            padding: 1rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 0.5rem;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }

        button:active {
            transform: translateY(0);
        }

        .divider {
            margin: 1.5rem 0;
            text-align: center;
            position: relative;
        }

        .divider::before {
            content: "";
            position: absolute;
            left: 0;
            top: 50%;
            width: 100%;
            height: 1px;
            background: #e0e0e0;
        }

        .divider span {
            background: white;
            padding: 0 1rem;
            position: relative;
            color: #999;
            font-size: 0.9rem;
        }

        .discord-login {
            background: #5865F2;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
        }

        .discord-login:hover {
            box-shadow: 0 10px 20px rgba(88, 101, 242, 0.4);
        }

        .error {
            background: #fee;
            color: #c33;
            padding: 1rem;
            border-radius: 0.5rem;
            margin-bottom: 1rem;
            display: none;
        }

        .footer {
            margin-top: 2rem;
            font-size: 0.85rem;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">🔐</div>
        <h1>GoBE Login</h1>
        <p>Access Kubex Ecosystem Dashboard</p>

        <div id="error" class="error"></div>

        <form class="login-form" onsubmit="handleLogin(event)">
            <input type="text" id="username" placeholder="Username" required>
            <input type="password" id="password" placeholder="Password" required>
            <button type="submit">Sign In</button>
        </form>

        <div class="divider"><span>OR</span></div>

        <button class="discord-login" onclick="loginWithDiscord()">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="white">
                <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515a.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0a12.64 12.64 0 0 0-.617-1.25a.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057a19.9 19.9 0 0 0 5.993 3.03a.078.078 0 0 0 .084-.028a14.09 14.09 0 0 0 1.226-1.994a.076.076 0 0 0-.041-.106a13.107 13.107 0 0 1-1.872-.892a.077.077 0 0 1-.008-.128a10.2 10.2 0 0 0 .372-.292a.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127a12.299 12.299 0 0 1-1.873.892a.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028a19.839 19.839 0 0 0 6.002-3.03a.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/>
            </svg>
            Continue with Discord
        </button>

        <div class="footer">
            Powered by <strong>GoBE v1.3.5</strong><br>
            Part of the Kubex Ecosystem
        </div>
    </div>

    <script>
        async function handleLogin(e) {
            e.preventDefault();

            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('error');

            try {
                const response = await fetch('/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password })
                });

                const data = await response.json();

                if (response.ok) {
                    // Store token in cookie
                    document.cookie = ` + "`gobe_token=${data.token}; path=/; max-age=86400`" + `;

                    // Redirect to original page or dashboard
                    const redirect = new URLSearchParams(window.location.search).get('redirect') || '/web';
                    window.location.href = redirect;
                } else {
                    errorDiv.textContent = data.message || 'Login failed';
                    errorDiv.style.display = 'block';
                }
            } catch (error) {
                errorDiv.textContent = 'Network error. Please try again.';
                errorDiv.style.display = 'block';
            }
        }

        function loginWithDiscord() {
            // Redirect to Discord OAuth
            const redirect = encodeURIComponent(window.location.origin + '/auth/discord/callback');
            window.location.href = '/auth/discord?redirect=' + redirect;
        }
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GenerateToken generates a new JWT token for a user
func (m *WebAuthMiddleware) GenerateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	if m.privateKey != nil { // pragma: allowlist secret
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(m.privateKey) // pragma: allowlist secret
	}

	if m.config.JWTSecret == "" {
		return "", fmt.Errorf("no signing credentials configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.JWTSecret)) // pragma: allowlist secret
}
