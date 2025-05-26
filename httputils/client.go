package httputils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/arniva/arniva-cloud-back-core/utils"
	"github.com/gofiber/fiber/v2"
)

/*
*	Option struct'ı, HTTP isteklerinde kullanılan parametreleri tutar.
*	Headers: HTTP isteğinde kullanılacak başlık bilgilerini tutar.
*	Query: HTTP isteğinde kullanılacak sorgu parametrelerini tutar.
*	Timeout: HTTP isteğinin zaman aşım süresini belirler.

* Default değerler:
*	DEFAULT_TIMEOUT: 30 saniye
*	DEFAULT_HEADERS:
		- Content-Type: application/json
		- Service: config.Cfg.AppName


*	HTTPResponse struct'ı, HTTP isteğine verilen yanıtı tutar.
*	StatusCode: HTTP yanıt durum kodunu tutar.
*	Body: HTTP yanıt içeriğini tutar.




*/

// var (
// 	DEFAULT_TIMEOUT = 30 * time.Second
// 	DEFAULT_HEADERS = map[string]string{
// 		"Content-Type": "application/json",
// 		"Service":      config.Cfg.AppName + " Service",
// 	}
// )

type Option struct {
	Headers map[string]string
	Query   map[string]string
	Timeout time.Duration
}

type HTTPResponse[T any] struct {
	StatusCode int
	Body       T
	Headers    map[string][]string
}

// sendRequest genel HTTP istek fonksiyonu
func sendRequest[T any](method, url string, body interface{}, option Option) (*HTTPResponse[T], error) {
	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	req := agent.Request()
	req.Header.SetMethod(method)
	req.SetRequestURI(url)

	if err := setRequestOptions(agent, option); err != nil {
		return nil, fmt.Errorf("request options ayarlanırken hata: %v", err)
	}

	if body != nil {
		agent.JSON(body)

	}

	if err := agent.Parse(); err != nil {
		return nil, fmt.Errorf("istek parse edilirken hata: %v", err)
	}

	code, respBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("istek gönderilirken hata: %v", errs[0])
	}

	if code >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", code, string(respBody))
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("yanıt parse edilirken hata: %v", err)
	}

	var HTTPResponse = HTTPResponse[T]{
		StatusCode: code,
		Body:       result,
		Headers:    nil,
	}

	return &HTTPResponse, nil
}

func Get[T any](url string, option Option) (*HTTPResponse[T], error) {
	return sendRequest[T](fiber.MethodGet, url, nil, option)
}

func Post[T any](url string, body interface{}, option Option) (*HTTPResponse[T], error) {
	return sendRequest[T](fiber.MethodPost, url, body, option)
}

func Put[T any](url string, body interface{}, option Option) (*HTTPResponse[T], error) {
	return sendRequest[T](fiber.MethodPut, url, body, option)
}

func Delete[T any](url string, option Option) (*HTTPResponse[T], error) {
	return sendRequest[T](fiber.MethodDelete, url, nil, option)
}

var (
	DEFAULT_TIMEOUT = 30 * time.Second
	DEFAULT_HEADERS = map[string]string{
		"Content-Type": "application/json",
		"Service":      utils.GetEnv("APP_NAME", "undefined"),
	}
)

func setRequestOptions(agent *fiber.Agent, option Option) error {

	if option.Timeout == 0 {
		option.Timeout = DEFAULT_TIMEOUT
	}
	agent.Timeout(option.Timeout)

	// First set default headers
	for key, value := range DEFAULT_HEADERS {
		agent.Set(key, value)
	}

	// Then set custom headers
	for key, value := range option.Headers {
		agent.Set(key, value)
	}

	if len(option.Query) > 0 {
		queryStr := ""
		for key, value := range option.Query {
			if queryStr != "" {
				queryStr += "&"
			}
			queryStr += fmt.Sprintf("%s=%s", key, value)
		}
		agent.QueryString(queryStr)
	}

	return nil
}
