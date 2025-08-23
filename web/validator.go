package web

import (
	"github.com/go-playground/validator/v10"

	"regexp"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func SanitizeInput(input string) string {
	re := regexp.MustCompile(`[^\w\s]`)
	return re.ReplaceAllString(input, "")
}
