package httpx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

const (
	CodeInvalidRequestBody = "INVALID_REQUEST_BODY"
	CodeValidationFailed   = "VALIDATION_FAILED"

	MsgInvalidRequestBody = "Request body is invalid"
	MsgValidationFailed   = "Request validation failed"
)

type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

type AppError struct {
	StatusCode int          `json:"-"`
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	Fields     []FieldError `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func BindAndValidate(c *fiber.Ctx, dst any) error {
	if dst == nil {
		return invalidRequestBodyError()
	}

	if err := c.BodyParser(dst); err != nil {
		return invalidRequestBodyError()
	}

	if err := validate.Struct(dst); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			fields := make([]FieldError, 0, len(validationErrors))
			for _, item := range validationErrors {
				fieldName := jsonFieldName(dst, item.StructField())
				fields = append(fields, FieldError{
					Field:   fieldName,
					Rule:    item.Tag(),
					Param:   item.Param(),
					Message: validationMessage(fieldName, item),
				})
			}

			return &AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       CodeValidationFailed,
				Message:    MsgValidationFailed,
				Fields:     fields,
			}
		}
		return &AppError{
			StatusCode: fiber.StatusBadRequest,
			Code:       CodeValidationFailed,
			Message:    MsgValidationFailed,
		}
	}
	return nil
}

func invalidRequestBodyError() *AppError {
	return &AppError{
		StatusCode: fiber.StatusBadRequest,
		Code:       CodeInvalidRequestBody,
		Message:    MsgInvalidRequestBody,
	}
}

func jsonFieldName(dst any, structField string) string {
	t := reflect.TypeOf(dst)
	if t == nil {
		return strings.ToLower(structField)
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}

	f, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}

	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(structField)
	}

	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(structField)
	}

	return name
}

func validationMessage(field string, err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at max %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
