package routes

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	"github.com/spf13/viper"
)

func SecureServerInit(r *gin.Engine, fullBindAddress string) error {
	trustedProxies, trustedProxiesErr := getTrustedProxies()
	if trustedProxiesErr != nil {
		return trustedProxiesErr
	}
	setTrustProxiesErr := r.SetTrustedProxies(trustedProxies)
	if setTrustProxiesErr != nil {
		return setTrustProxiesErr
	}

	r.Use(
		func(c *gin.Context) {
			if !validateExpectedHosts(fullBindAddress, c) {
				c.Abort()
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

				// Handle OPTIONS preflight requests
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusOK)
					return
				}

				c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
				c.Header("Referrer-Policy", "strict-origin")
				c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
				c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")

				// c.Header("X-Frame-Options", "DENY")
				c.Header("X-XSS-Protection", "1; mode=block")
				c.Header("X-Content-Type-Options", "nosniff")

				c.Next()
			}
		},
	)

	return nil
}

func getTrustedProxies() ([]string, error) {
	trustedProxies := viper.GetStringSlice("trustedProxies")
	if len(trustedProxies) == 0 {
		interfaces, err := net.Interfaces()
		if err != nil {
			return []string{}, err
		}

		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback == 0 {
				addrs, addrsErr := iface.Addrs()
				if addrsErr != nil {
					return []string{}, fmt.Errorf("error getting addresses for interface %s: %s", iface.Name, addrsErr)
					//continue // Ignora erro
				}

				for _, addr := range addrs {
					ipNet, ok := addr.(*net.IPNet)
					if ok {
						trustedProxies = append(trustedProxies, ipNet.IP.String())
					}
				}
			}
		}
	}

	gl.Log("notice", "Trusted Proxies: %v", trustedProxies)

	return trustedProxies, nil
}

func validateExpectedHosts(fullBindAddress string, c *gin.Context) bool {
	// if c.Request.Host == fullBindAddress ||
	// 	c.Request.URL.Host == fullBindAddress {
	// 	return true
	// }

	// bindPort := strings.Split(fullBindAddress, ":")[1]
	// trustedLocalList := []string{"localhost", "127.0.0.1", "localhost:" + bindPort, "127.0.0.1:" + bindPort}
	// for _, trustedLocal := range trustedLocalList {
	// 	if c.Request.Host == trustedLocal ||
	// 		c.Request.URL.Host == trustedLocal {
	// 		return true
	// 	}
	// }

	//c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Unauthorized host: " + c.Request.Host})
	// return false

	return true
}

func GetDefaultRouteMap(rtr ci.IRouter) map[string]map[string]ci.IRoute {
	return map[string]map[string]ci.IRoute{
		"cronRoutes":           NewCronRoutes(&rtr),
		"webhookRoutes":        NewWebhookRoutes(&rtr),
		"contactRoutes":        NewContactRoutes(&rtr),
		"authRoutes":           NewAuthRoutes(&rtr),
		"userRoutes":           NewUserRoutes(&rtr),
		"productRoutes":        NewProductRoutes(&rtr),
		"serverRoutes":         NewServerRoutes(&rtr),
		"customerRoutes":       NewCustomerRoutes(&rtr),
		"discordRoutes":        NewDiscordRoutes(&rtr),
		"mcpTasksRoutes":       NewMCPTasksRoutes(&rtr),
		"mcpProvidersRoutes":   NewMCPProvidersRoutes(&rtr),
		"mcpLLMRoutes":         NewMCPLLMRoutes(&rtr),
		"mcpPreferencesRoutes": NewMCPPreferencesRoutes(&rtr),
	}
}

func uniqueMiddlewareStack(middlewares []gin.HandlerFunc) []gin.HandlerFunc {
	uniqueMap := make(map[string]gin.HandlerFunc)
	uniqueList := []gin.HandlerFunc{}

	for _, middleware := range middlewares {
		funcPtr := fmt.Sprintf("%p", middleware) // Obtém o endereço da função como string

		if _, exists := uniqueMap[funcPtr]; !exists {
			uniqueMap[funcPtr] = middleware
			uniqueList = append(uniqueList, middleware)
		}
	}

	return uniqueList
}
