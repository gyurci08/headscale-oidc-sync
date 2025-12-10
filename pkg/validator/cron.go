package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func registerCronValidator() {
	Validate.RegisterValidation("cron", func(fl validator.FieldLevel) bool {
		cronRegex := `^(@every \d+[smhdw])|(@hourly)|(@daily)|(@weekly)|(@monthly)|(@yearly)$`
		value := fl.Field().String()
		matched, _ := regexp.MatchString(cronRegex, value)
		return matched
	})
}
