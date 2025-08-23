package middlewares

import (
	"github.com/gin-gonic/gin"
)

func ValidateAndSanitize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var input map[string]interface{}
		// if err := c.ShouldBindJSON(&input); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		// 	c.Abort()
		// 	return
		// }

		// for key, value := range input {
		// 	if str, ok := value.(string); ok {
		// 		input[key] = sanitizedInput(str)
		// 	}
		// }

		// if err := validateStruct(input); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	c.Abort()
		// 	return
		// }

		// c.Set("sanitizedInput", input)
		c.Next()
	}
}

func ValidateAndSanitizeBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var input map[string]interface{}
		// if err := c.ShouldBindJSON(&input); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		// 	c.Abort()
		// 	return
		// }

		// for key, value := range input {
		// 	if str, ok := value.(string); ok {
		// 		input[key] = sanitizedInput(str)
		// 	}
		// }

		// if err := validateStruct(input); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	c.Abort()
		// 	return
		// }

		// c.Set("sanitizedBody", input)
		c.Next()
	}
}

// func sanitizedInput(input string) string {
// 	// Implement your sanitization logic here
// 	// For example, removing HTML tags, trimming whitespace, etc.
// 	return input // Placeholder for actual sanitization logic
// }

// func validateStruct(input map[string]interface{}) error {
// 	// Implement your validation logic here
// 	// For example, checking required fields, field types, etc.
// 	return nil // Placeholder for actual validation logic
// }
