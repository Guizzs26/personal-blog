package validatorx

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func ValidateStruct(input any) error {
	err := validate.Struct(input)
	if err == nil {
		return nil
	}

	if _, ok := err.(validator.ValidationErrors); ok {
		return err
	}

	return nil
}

func FormatValidationErrors(errs validator.ValidationErrors) map[string]string {
	formatted := make(map[string]string)

	for _, err := range errs {
		field := toSnakeCase(err.Field())

		var msg string
		switch err.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", field)
		case "uuidv4":
			msg = fmt.Sprintf("%s must be a valid UUIDV4", field)
		default:
			msg = fmt.Sprintf("%s is invalid", field)
		}
		formatted[field] = msg
	}
	return formatted
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	runes := []rune(s)

	for i := range runes {
		if i > 0 && isUpper(runes[i]) && (isLower(runes[i-1]) || (i+1 < len(runes) && isLower(runes[i+1]))) {
			sb.WriteByte('_')
		}
		sb.WriteRune(runes[i])
	}
	return strings.ToLower(sb.String())
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}
