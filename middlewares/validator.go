package middlewares

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/arniva/arniva-cloud-back-core/httputils"
	"github.com/go-playground/locales/tr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	trTranslations "github.com/go-playground/validator/v10/translations/tr"
	"github.com/gofiber/fiber/v2"
)

type XValidator struct {
	Validator *validator.Validate
}

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
	Str         string
}

var Validator = &XValidator{Validator: validator.New()}
var turkish = tr.New()
var uni = ut.New(turkish, turkish)
var translator, f = uni.GetTranslator("tr")

func init() {

	if err := trTranslations.RegisterDefaultTranslations(Validator.Validator, translator); err != nil {
		panic(fmt.Sprintf("Failed to register translations: %v", err))
	}

}

func (x *XValidator) Validate(data interface{}, exclude *[]string) []ErrorResponse {
	// Test: çeviri kayıtlı mı?
	// fmt.Println(translator.T("required", "StokID")) // "StokID is a required field"

	validationErrors := []ErrorResponse{}

	excludeMap := make(map[string]interface{})
	for _, v := range *exclude {
		excludeMap[strings.ToLower(v)] = nil
	}

	val := reflect.ValueOf(data)
	kind := val.Kind()
	if kind == reflect.Ptr {
		val = val.Elem()
		kind = val.Kind()
	}

	if kind == reflect.Slice || kind == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			errs := x.Validator.Struct(item)
			if errs != nil {
				for _, err := range errs.(validator.ValidationErrors) {
					var elem ErrorResponse
					elem.FailedField = err.Field()
					elem.Tag = err.Tag()
					elem.Value = err.Value()
					elem.Str = err.Translate(translator)
					elem.Error = true
					if _, exist := excludeMap[elem.FailedField]; !exist {
						validationErrors = append(validationErrors, elem)
					}
				}
			}
		}
	} else {
		errs := x.Validator.Struct(data)

		if errs != nil {
			for _, err := range errs.(validator.ValidationErrors) {
				// In this case data object is actually holding the User struct
				var elem ErrorResponse

				elem.FailedField = err.Field()       // Export struct field name
				elem.Tag = err.Tag()                 // Export struct tag
				elem.Value = err.Value()             // Export field value
				elem.Str = err.Translate(translator) // Export translated error message
				elem.Error = true

				if _, exist := excludeMap[elem.FailedField]; !exist {
					validationErrors = append(validationErrors, elem)
				}
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
				errMsgs[i] = err.Str
			}
			return httputils.NewApiError(400, httputils.Enums.Code2, strings.Join(errMsgs, " and "))
		}

		return c.Next()
	}
}

//
