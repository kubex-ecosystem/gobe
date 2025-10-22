package middlewares

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	gl "github.com/kubex-ecosystem/logz/logger"
)

const (
	DiscordSignatureHeader     = "X-Signature-Ed25519"
	DiscordTimestampHeader     = "X-Signature-Timestamp"
	DiscordEventIDHeader       = "X-Discord-Event-Id"
	DiscordSignatureContextKey = "discord.signature.raw"
	DiscordTimestampContextKey = "discord.signature.timestamp"
	DiscordEventIDContextKey   = "discord.signature.event_id"
	DiscordBodyContextKey      = "discord.signature.body"
	DiscordVerifiedContextKey  = "discord.signature.verified"
)

// DiscordWebhookGuard logs incoming signature headers and optionally verifies them when enabled in config.
// When verification is disabled (default), the middleware simply stores headers/body in the request context
// so controllers can react appropriately. Verification will run once VerifySignatures is true and a public key is available.
func DiscordWebhookGuard(cfg bootstrap.DiscordConfig) gin.HandlerFunc {
	verify := cfg.Features.VerifySignatures
	publicKeyHex := strings.TrimSpace(cfg.OAuth2.PublicKey)

	return func(c *gin.Context) {
		signature := strings.TrimSpace(c.GetHeader(DiscordSignatureHeader))
		timestamp := strings.TrimSpace(c.GetHeader(DiscordTimestampHeader))
		eventID := strings.TrimSpace(c.GetHeader(DiscordEventIDHeader))

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			gl.Log("error", "discord webhook guard: failed to read body", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Set(DiscordSignatureContextKey, signature)
		c.Set(DiscordTimestampContextKey, timestamp)
		c.Set(DiscordEventIDContextKey, eventID)
		c.Set(DiscordBodyContextKey, bodyBytes)

		if !verify {
			c.Set(DiscordVerifiedContextKey, false)
			gl.Log("debug", "discord webhook guard bypass (verify_signatures=false)")
			c.Next()
			return
		}

		if signature == "" || timestamp == "" {
			gl.Log("warn", "discord webhook guard: missing signature headers with verification enabled")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_signature"})
			return
		}
		if len(publicKeyHex) == 0 {
			gl.Log("warn", "discord webhook guard: verification enabled without public key, bypassing")
			c.Set(DiscordVerifiedContextKey, false)
			c.Next()
			return
		}

		pkBytes, err := hex.DecodeString(publicKeyHex)
		if err != nil {
			gl.Log("error", "discord webhook guard: invalid public key hex", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid_public_key"})
			return
		}
		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			gl.Log("warn", "discord webhook guard: invalid signature hex", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_signature"})
			return
		}

		payload := append([]byte(timestamp), bodyBytes...)
		if !ed25519.Verify(pkBytes, payload, sigBytes) {
			gl.Log("warn", "discord webhook guard: signature verification failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "signature_mismatch"})
			return
		}

		c.Set(DiscordVerifiedContextKey, true)
		c.Next()
	}
}
