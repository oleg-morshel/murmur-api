package core_http_request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

const maxRequestBodySize = 1 << 20

var requestValidator = validator.New()

func init() {
	requestValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

type validatable interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("%w: invalid request body", core_errors.ErrBadRequest)
	}

	if err := requestValidator.Struct(dest); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return fmt.Errorf("%w: %s", core_errors.ErrBadRequest, formatValidationErrors(validationErrors))
		}
		return fmt.Errorf("%w: validation failed", core_errors.ErrBadRequest)
	}

	if v, ok := dest.(validatable); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	messages := make([]string, 0, len(errs))

	for _, e := range errs {
		field := e.Field()
		switch e.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "email":
			messages = append(messages, fmt.Sprintf("%s must be a valid email", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		case "alphanum":
			messages = append(messages, fmt.Sprintf("%s must contain only letters and numbers", field))
		default:
			messages = append(messages, fmt.Sprintf("%s failed on %s", field, e.Tag()))
		}
	}

	return strings.Join(messages, "; ")
}
