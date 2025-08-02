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
		// 		input[key] = au.SanitizeInput(str)
		// 	}
		// }

		// if err := au.ValidateStruct(input); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	c.Abort()
		// 	return
		// }

		// c.Set("sanitizedInput", input)
		c.Next()
	}
}
