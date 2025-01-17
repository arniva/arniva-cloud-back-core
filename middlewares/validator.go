package middlewares

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/ismailozdel/core2/httputils"
)

type XValidator struct {
	Validator *validator.Validate
}

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

var Validator = &XValidator{Validator: validator.New()}

func (x *XValidator) Validate(data interface{}, exclude *[]string) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	excludeMap := make(map[string]interface{})
	for _, v := range *exclude {
		excludeMap[strings.ToLower(v)] = nil
	}

	errs := x.Validator.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			// In this case data object is actually holding the User struct
			var elem ErrorResponse

			elem.FailedField = err.Field() // Export struct field name
			elem.Tag = err.Tag()           // Export struct tag
			elem.Value = err.Value()       // Export field value
			elem.Error = true

			if _, exist := excludeMap[elem.FailedField]; !exist {
				validationErrors = append(validationErrors, elem)
			}
		}
	}

	return validationErrors
}

func Validate[T any](exclude []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := c.Locals("data").(*T)
		if errs := Validator.Validate(data, &exclude); len(errs) > 0 && errs[0].Error {
			errMsgs := make([]string, len(errs))
			for i, err := range errs {
				errMsgs[i] = fmt.Sprintf(
					"[%s]: '%v' | Needs to implement '%s'",
					err.FailedField,
					err.Value,
					err.Tag,
				)
			}
			return httputils.NewApiError(400, httputils.Enums.Code2, strings.Join(errMsgs, " and "))
		}

		return c.Next()
	}
}
