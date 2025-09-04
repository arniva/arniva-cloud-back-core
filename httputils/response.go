package httputils

import (
	"github.com/arniva/arniva-cloud-back-core/config"
	"github.com/gofiber/fiber/v2"
)

type packageEnums struct {
	ErrorInvalidRequest string
	ErrorUnauthorized   string
	ErrorInternalServer string
	Code0               int
	Code1               int
	Code2               int
	Code3               int
}

var Enums = packageEnums{
	ErrorInvalidRequest: "Geçersiz istek formatı",
	ErrorInternalServer: "İç sunucu hatası",
	ErrorUnauthorized:   "Yetkisiz istek",

	// Error kodları
	Code0: 0,
	Code1: 1,
	Code2: 2,
	Code3: 3,
}

// AppError özel hata yapısı
type ApiError struct {
	StatusCode int
	Message    string
	Code       int
}

func (e *ApiError) Error() string {
	return e.Message
}

// NewAppError yeni bir AppError oluşturur
func NewApiError(statusCode int, code int, message string) *ApiError {
	return &ApiError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// Response standart API yanıt yapısı
type Response struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	StatusCode int         `json:"-"`
	Meta       *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	TotalRecord   int64  `json:"total_records"`
	CurrentOffset int    `json:"current_offset"`
	Limit         int    `json:"limit"`
	Sort          string `json:"sort"`
}

// NewSuccessResponse başarılı yanıt oluşturur
func NewSuccessResponse(data interface{}, meta ...Meta) *Response {
	res := Response{
		StatusCode: 200,
		Message:    "",
		Data:       data,
		Code:       Enums.Code0,
		Meta:       nil,
	}

	var m Meta
	if len(meta) != 0 {
		m = meta[0]
		res = Response{
			StatusCode: 200,
			Message:    "",
			Data:       data,
			Code:       Enums.Code0,
			Meta:       &m,
		}
	}

	return &res
}

// NewErrorResponse hata yanıtı oluşturur
func NewErrorResponse(statusCode int, code int, message string) *Response {
	appName := config.Cfg.AppName + " | "
	return &Response{
		StatusCode: statusCode,
		Message:    appName + message,
		Data:       nil,
		Code:       code,
	}
}

// Send yanıtı Fiber context üzerinden gönderir
func (r *Response) Send(c *fiber.Ctx) error {
	return c.Status(r.StatusCode).JSON(r)
}

func PrepareNotFoundError(message string) *ApiError {
	return NewApiError(
		fiber.StatusNotFound,
		Enums.Code1,
		"Not Found"+"| "+message,
	)

}

func PrepareParseError(message string) *ApiError {
	return NewApiError(
		fiber.StatusBadRequest,
		Enums.Code2,
		Enums.ErrorInvalidRequest+"| "+message,
	)
}
func PrepareUnauthorizedRequestError(message string) *ApiError {
	return NewApiError(
		fiber.StatusUnauthorized,
		1,
		Enums.ErrorUnauthorized+" | "+message,
	)
}

func PrepareInternalServerError(message string) *ApiError {
	return NewApiError(
		fiber.StatusInternalServerError,
		2,
		Enums.ErrorInternalServer+"| "+message,
	)
}
