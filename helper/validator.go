package helper

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
	})
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
				r != '.' && r != '_' {
				return false
			}
		}
		return true
	})
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return checkPasswordStrength(fl.Field().String()) == ""
	})

	return v
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "invalid object for validation"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validation failed"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
}

const minPasswordLength = 8

func checkPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "must be at least 8 characters"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "must contain both letters and numbers"
	}

	weak := map[string]bool{
		"password1": true, "12345678": true, "qwerty123": true,
		"admin123": true, "password123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password is too common"
	}

	return ""
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "invalid email format"
	case "min":
		if fe.Kind() == reflect.String {
			return "must be at least " + fe.Param() + " characters"
		}
		return "must be at least " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "must be at most " + fe.Param() + " characters"
		}
		return "must be at most " + fe.Param()
	case "alphanum":
		return "must only contain letters and numbers"
	case "number":
		return "must only contain numbers"
	case "nospace":
		return "cannot contain spaces"
	case "username":
		return "can only contain letters, numbers, periods, and underscores"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return checkPasswordStrength(value)
		}
		return "password does not meet requirements"
	case "oneof":
		return "must be one of: " +
			strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "failed rule " + fe.Tag()
	}
}
