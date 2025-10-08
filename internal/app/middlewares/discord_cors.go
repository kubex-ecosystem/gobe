// Package middlewares provides CORS configuration for Discord iframe embedding
package middlewares

import (
	"github.com/gin-gonic/gin"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// DiscordCORSMiddleware configures CORS for Discord Activity/iframe embedding
func DiscordCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow Discord domains for iframe embedding
		allowedOrigins := []string{
			"https://discord.com",
			"https://ptb.discord.com",    // Public Test Build
			"https://canary.discord.com", // Canary
			"https://discordapp.com",     // Legacy
			"http://localhost:3000",      // Local development
			"http://localhost:3666",      // GoBE local
		}

		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			// Set CORS headers for Discord
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Accept, Authorization, Content-Type, X-CSRF-Token, X-Auth-Token, X-Session-ID, X-Discord-User-ID, X-Discord-Guild-ID")
			c.Writer.Header().Set("Access-Control-Expose-Headers",
				"Authorization, X-Auth-Token, X-Session-ID")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Discord-specific headers for iframe embedding
			c.Writer.Header().Set("X-Frame-Options", "ALLOW-FROM https://discord.com")
			c.Writer.Header().Set("Content-Security-Policy",
				"frame-ancestors 'self' https://discord.com https://*.discord.com https://discordapp.com")

			gl.Log("debug", "Discord CORS applied", "origin", origin)
		} else {
			// Default restrictive CORS for non-Discord origins
			c.Writer.Header().Set("Access-Control-Allow-Origin", "https://gobe.kubex.io")
			c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ExtractDiscordAuth extracts Discord-specific auth headers
func ExtractDiscordAuth(c *gin.Context) (userID string, guildID string) {
	// Discord passes user/guild info via custom headers in Activities
	userID = c.GetHeader("X-Discord-User-ID")
	guildID = c.GetHeader("X-Discord-Guild-ID")

	// Fallback to query parameters (for initial OAuth flow)
	if userID == "" {
		userID = c.Query("discord_user_id")
	}
	if guildID == "" {
		guildID = c.Query("discord_guild_id")
	}

	return userID, guildID
}

// ValidateDiscordOrigin checks if request comes from Discord
func ValidateDiscordOrigin(c *gin.Context) bool {
	origin := c.Request.Header.Get("Origin")
	referer := c.Request.Header.Get("Referer")

	discordDomains := []string{
		"discord.com",
		"ptb.discord.com",
		"canary.discord.com",
		"discordapp.com",
	}

	for _, domain := range discordDomains {
		if origin == "https://"+domain ||
			(referer != "" && contains(referer, domain)) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
