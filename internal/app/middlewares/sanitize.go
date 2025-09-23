package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	// HTML tags regex for sanitization
	htmlTagsRegex = regexp.MustCompile(`<[^>]*>`)
	// SQL injection patterns
	sqlInjectionRegex = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
	// XSS patterns
	xssRegex = regexp.MustCompile(`(?i)(javascript:|data:|vbscript:|on\w+\s*=)`)
	// Path traversal patterns
	pathTraversalRegex = regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`)
	// Command injection patterns
	cmdInjectionRegex = regexp.MustCompile(`[;&|` + "`" + `$()]`)

	// Global validator instance
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}

// ValidateAndSanitize provides comprehensive input sanitization for general purposes
func ValidateAndSanitize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		sanitizeQueryParams(c)

		// Sanitize URL path parameters
		sanitizePathParams(c)

		// Sanitize headers (specific security-sensitive ones)
		sanitizeHeaders(c)

		c.Next()
	}
}

// ValidateAndSanitizeBody provides comprehensive body sanitization and validation
func ValidateAndSanitizeBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input map[string]interface{}

		// Try to bind JSON body
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON input",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Recursively sanitize all string values in the input
		sanitizedInput, err := sanitizeMapRecursive(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Input sanitization failed",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate the sanitized input structure
		if err := validateStruct(sanitizedInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Input validation failed",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Store sanitized body for use in handlers
		c.Set("sanitizedBody", sanitizedInput)
		c.Next()
	}
}

// sanitizeQueryParams sanitizes all query parameters
func sanitizeQueryParams(c *gin.Context) {
	for key, values := range c.Request.URL.Query() {
		for i, value := range values {
			values[i] = sanitizeInput(value)
		}
		c.Request.URL.Query()[key] = values
	}
}

// sanitizePathParams sanitizes URL path parameters
func sanitizePathParams(c *gin.Context) {
	for _, param := range c.Params {
		param.Value = sanitizeInput(param.Value)
	}
}

// sanitizeHeaders sanitizes security-sensitive headers
func sanitizeHeaders(c *gin.Context) {
	headersToSanitize := []string{
		"User-Agent",
		"Referer",
		"Origin",
		"X-Forwarded-For",
		"X-Real-IP",
	}

	for _, header := range headersToSanitize {
		if value := c.GetHeader(header); value != "" {
			c.Request.Header.Set(header, sanitizeInput(value))
		}
	}
}

// sanitizeMapRecursive recursively sanitizes all string values in a map
func sanitizeMapRecursive(input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range input {
		// Sanitize the key itself
		sanitizedKey := sanitizeInput(key)

		switch v := value.(type) {
		case string:
			// Sanitize string values
			result[sanitizedKey] = sanitizeInput(v)
		case map[string]interface{}:
			// Recursively sanitize nested maps
			sanitizedNested, err := sanitizeMapRecursive(v)
			if err != nil {
				return nil, fmt.Errorf("failed to sanitize nested map for key %s: %w", key, err)
			}
			result[sanitizedKey] = sanitizedNested
		case []interface{}:
			// Sanitize array elements
			sanitizedArray, err := sanitizeArrayRecursive(v)
			if err != nil {
				return nil, fmt.Errorf("failed to sanitize array for key %s: %w", key, err)
			}
			result[sanitizedKey] = sanitizedArray
		default:
			// Keep other types as-is (numbers, booleans, etc.)
			result[sanitizedKey] = value
		}
	}

	return result, nil
}

// sanitizeArrayRecursive recursively sanitizes array elements
func sanitizeArrayRecursive(input []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(input))

	for i, value := range input {
		switch v := value.(type) {
		case string:
			result[i] = sanitizeInput(v)
		case map[string]interface{}:
			sanitizedMap, err := sanitizeMapRecursive(v)
			if err != nil {
				return nil, fmt.Errorf("failed to sanitize map in array at index %d: %w", i, err)
			}
			result[i] = sanitizedMap
		case []interface{}:
			sanitizedArray, err := sanitizeArrayRecursive(v)
			if err != nil {
				return nil, fmt.Errorf("failed to sanitize nested array at index %d: %w", i, err)
			}
			result[i] = sanitizedArray
		default:
			result[i] = value
		}
	}

	return result, nil
}

// sanitizeInput performs comprehensive input sanitization
func sanitizeInput(input string) string {
	if input == "" {
		return input
	}

	// 1. Trim whitespace
	sanitized := strings.TrimSpace(input)

	// 2. HTML decode to handle encoded entities
	sanitized = html.UnescapeString(sanitized)

	// 3. URL decode to handle URL-encoded payloads
	if decoded, err := url.QueryUnescape(sanitized); err == nil {
		sanitized = decoded
	}

	// 4. Remove HTML tags
	sanitized = htmlTagsRegex.ReplaceAllString(sanitized, "")

	// 5. Remove dangerous patterns
	sanitized = removeDangerousPatterns(sanitized)

	// 6. Remove control characters (except common ones like \n, \t, \r)
	sanitized = removeControlCharacters(sanitized)

	// 7. HTML escape the result for safe output
	sanitized = html.EscapeString(sanitized)

	// 8. Limit length to prevent buffer overflow attacks
	if len(sanitized) > 10000 {
		sanitized = sanitized[:10000]
	}

	return sanitized
}

// removeDangerousPatterns removes known dangerous patterns
func removeDangerousPatterns(input string) string {
	// Remove SQL injection patterns
	input = sqlInjectionRegex.ReplaceAllString(input, "")

	// Remove XSS patterns
	input = xssRegex.ReplaceAllString(input, "")

	// Remove path traversal patterns
	input = pathTraversalRegex.ReplaceAllString(input, "")

	// Remove command injection patterns
	input = cmdInjectionRegex.ReplaceAllString(input, "")

	return input
}

// removeControlCharacters removes dangerous control characters
func removeControlCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		// Keep printable characters and common whitespace
		if unicode.IsPrint(r) || r == '\n' || r == '\t' || r == '\r' || r == ' ' {
			return r
		}
		// Remove other control characters
		return -1
	}, input)
}

// validateStruct validates the structure and content of input data
func validateStruct(input map[string]interface{}) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	// Check for empty input
	if len(input) == 0 {
		return errors.New("input cannot be empty")
	}

	// Validate each field recursively
	for key, value := range input {
		if err := validateField(key, value); err != nil {
			return fmt.Errorf("validation failed for field '%s': %w", key, err)
		}
	}

	return nil
}

// validateField validates individual fields
func validateField(key string, value interface{}) error {
	// Validate key format
	if strings.TrimSpace(key) == "" {
		return errors.New("field key cannot be empty")
	}

	// Check for suspicious key patterns
	if sqlInjectionRegex.MatchString(key) || xssRegex.MatchString(key) {
		return fmt.Errorf("field key contains suspicious patterns: %s", key)
	}

	// Validate based on value type
	switch v := value.(type) {
	case string:
		return validateStringField(key, v)
	case map[string]interface{}:
		return validateStruct(v)
	case []interface{}:
		return validateArrayField(key, v)
	case nil:
		// Nil values are generally acceptable
		return nil
	default:
		// Validate using reflection for structured data
		return validateWithReflection(value)
	}
}

// validateStringField validates string fields
func validateStringField(key, value string) error {
	// Check length limits
	if len(value) > 10000 {
		return fmt.Errorf("string field '%s' exceeds maximum length of 10000 characters", key)
	}

	// Check for null bytes
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("string field '%s' contains null bytes", key)
	}

	// Additional validation for specific field patterns
	switch {
	case strings.Contains(strings.ToLower(key), "email"):
		return validateEmail(value)
	case strings.Contains(strings.ToLower(key), "url"):
		return validateURL(value)
	case strings.Contains(strings.ToLower(key), "phone"):
		return validatePhone(value)
	}

	return nil
}

// validateArrayField validates array fields
func validateArrayField(key string, values []interface{}) error {
	// Check array size limits
	if len(values) > 1000 {
		return fmt.Errorf("array field '%s' exceeds maximum length of 1000 items", key)
	}

	// Validate each array element
	for i, value := range values {
		if err := validateField(fmt.Sprintf("%s[%d]", key, i), value); err != nil {
			return err
		}
	}

	return nil
}

// validateWithReflection validates structured data using reflection
func validateWithReflection(value interface{}) error {
	if value == nil {
		return nil
	}

	// Convert to JSON and back to ensure it's serializable
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("value is not JSON serializable: %w", err)
	}

	// Check JSON size
	if len(jsonData) > 100000 {
		return errors.New("structured data exceeds maximum size of 100KB")
	}

	return nil
}

// Email validation
func validateEmail(email string) error {
	if email == "" {
		return nil // Empty emails are handled at business logic level
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// URL validation
func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL must use http or https scheme")
	}

	return nil
}

// Phone validation (basic international format)
func validatePhone(phone string) error {
	if phone == "" {
		return nil
	}

	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	if !phoneRegex.MatchString(cleanPhone) {
		return errors.New("invalid phone number format")
	}

	return nil
}
