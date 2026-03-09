package handlers

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateRequest validates a request struct using validator
func ValidateRequest(req interface{}) error {
	if err := validate.Struct(req); err != nil {
		return err
	}
	return nil
}
