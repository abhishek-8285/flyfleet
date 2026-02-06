package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateField(field interface{}, tag string) error
	RegisterValidation(tag string, fn validator.Func) error
}

type StructValidator struct {
	validate *validator.Validate
}

func New() (Validator, error) {
	v := validator.New()
	
	// Register custom validations
	if err := registerCustomValidations(v); err != nil {
		return nil, err
	}
	
	// Register field name translations
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &StructValidator{validate: v}, nil
}

func (sv *StructValidator) ValidateStruct(s interface{}) error {
	return sv.validate.Struct(s)
}

func (sv *StructValidator) ValidateField(field interface{}, tag string) error {
	return sv.validate.Var(field, tag)
}

func (sv *StructValidator) RegisterValidation(tag string, fn validator.Func) error {
	return sv.validate.RegisterValidation(tag, fn)
}

func registerCustomValidations(v *validator.Validate) error {
	// Phone number validation (Indian format)
	if err := v.RegisterValidation("phone", phoneValidator); err != nil {
		return err
	}
	
	// License plate validation
	if err := v.RegisterValidation("license_plate", licensePlateValidator); err != nil {
		return err
	}
	
	// UUID validation
	if err := v.RegisterValidation("uuid", uuidValidator); err != nil {
		return err
	}
	
	// Future date validation
	if err := v.RegisterValidation("future_date", futureDateValidator); err != nil {
		return err
	}
	
	return nil
}

func phoneValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Allow empty values, use required tag separately
	}
	
	// Indian mobile number pattern
	pattern := `^[6-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func licensePlateValidator(fl validator.FieldLevel) bool {
	plate := fl.Field().String()
	if plate == "" {
		return true
	}
	
	// Indian license plate format: XX00XX0000
	pattern := `^[A-Z]{2}[0-9]{2}[A-Z]{2}[0-9]{4}$`
	matched, _ := regexp.MatchString(pattern, plate)
	return matched
}

func uuidValidator(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()
	if uuid == "" {
		return true
	}
	
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(uuid))
	return matched
}

func futureDateValidator(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	
	return date.After(time.Now())
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

// ValidationErrors wraps multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	
	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	
	return strings.Join(messages, "; ")
}

// ProcessValidationError converts validator errors to our format
func ProcessValidationError(err error) error {
	if err == nil {
		return nil
	}
	
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	
	errors := make(ValidationErrors, len(validationErrs))
	for i, fieldErr := range validationErrs {
		errors[i] = ValidationError{
			Field:   fieldErr.Field(),
			Value:   fieldErr.Value(),
			Tag:     fieldErr.Tag(),
			Param:   fieldErr.Param(),
			Message: buildErrorMessage(fieldErr),
		}
	}
	
	return errors
}

func buildErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "phone":
		return "Invalid phone number format"
	case "license_plate":
		return "Invalid license plate format"
	case "uuid":
		return "Invalid UUID format"
	case "min":
		return fmt.Sprintf("Minimum length is %s", fe.Param())
	case "max":
		return fmt.Sprintf("Maximum length is %s", fe.Param())
	case "gte":
		return fmt.Sprintf("Value must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Value must be less than or equal to %s", fe.Param())
	case "future_date":
		return "Date must be in the future"
	default:
		return fmt.Sprintf("Validation failed on '%s' tag", fe.Tag())
	}
}