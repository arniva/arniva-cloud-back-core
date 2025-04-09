package httputils

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func BodyParser[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		if err := c.BodyParser(&data); err != nil {
			log.Print(err)
			return PrepareParseError(err.Error())
		}
		c.Locals("data", &data)

		return c.Next()
	}
}

var operators = map[string]string{
	"eq":  "=", // eşittir
	"ne":  "<>",
	"co":  "ILIKE",
	"sw":  "ILIKE",
	"ew":  "ILIKE",
	"gt":  ">",
	"ge":  ">=",
	"lt":  "<",
	"le":  "<=",
	"or":  "OR",
	"and": "AND",
	"in":  "IN",
	"nin": "NOT IN",
	"is":  "IS",
	"nis": "IS NOT",
}

// return sql query and values
func ParseQueryToSql[T any](query string) (string, []interface{}, error) {
	// example /Messages?filter=( action eq LOGIN and messageType ne IVR ) and messageId sw 2d0
	parts := strings.Split(query, " ")
	sqlParts := make([]string, 0)
	values := make([]interface{}, 0)
	sqlParts = append(sqlParts, "1=1 AND")
	var prevOperator string
	allowedFields := extractAllowedFields[T]()
	// {key: tip}
	for _, part := range parts {

		if part == "(" || part == ")" {
			sqlParts = append(sqlParts, part)
			continue
		}

		if op, ok := operators[part]; ok {
			if op == "OR" || op == "AND" {
				sqlParts = append(sqlParts, op)
				continue
			}
			sqlParts = append(sqlParts, op)
			prevOperator = part
			continue
		}
		if prevOperator == "" || prevOperator == "OR" || prevOperator == "AND" || prevOperator == "or" || prevOperator == "and" {
			if _, ok := allowedFields[part]; !ok {
				return "", nil, fmt.Errorf("Invalid key: %s", part)
			}
			sqlParts = append(sqlParts, part)
			continue
		}

		if prevOperator == "co" {
			sqlParts = append(sqlParts, "?")
			values = append(values, "%"+part+"%")
		} else if prevOperator == "sw" {
			sqlParts = append(sqlParts, "?")
			values = append(values, part+"%")
		} else if prevOperator == "ew" {
			sqlParts = append(sqlParts, "?")
			values = append(values, "%"+part)
		} else if prevOperator == "in" || prevOperator == "nin" {
			sqlParts = append(sqlParts, "(?)")
			values = append(values, strings.Split(part, "|"))
		} else {
			//( action eq LOGIN and messageType ne IVR ) and messageId sw 2d0
			sqlParts = append(sqlParts, "?")
			values = append(values, part)
		}

		prevOperator = ""
	}
	if sqlParts[len(sqlParts)-1] == "AND" || sqlParts[len(sqlParts)-1] == "OR" {
		sqlParts = sqlParts[:len(sqlParts)-1]
	}
	return strings.Join(sqlParts, " "), values, nil
}

func extractAllowedFields[T any]() map[string]string {
	var model T
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	fields := map[string]string{"limit": "int", "offset": "int"} // Limit ve Offset default
	extractFields(modelType, fields)
	return fields
}

func extractFields(modelType reflect.Type, fields map[string]string) {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Anonymous {
			// If the field is an embedded struct, recursively extract its fields
			extractFields(field.Type, fields)
		} else {
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Name
			}
			fields[jsonTag] = field.Type.String()
		}
	}
}

func parseKey(k string) (string, string) {
	validOps := []string{"<>", ">", "<"}
	for _, op := range validOps {
		if strings.HasPrefix(k, op) {
			return strings.TrimPrefix(k, op), op
		}
	}
	return k, "="
}

func buildSql(fieldName, fieldType, operator string, value interface{}) (string, interface{}) {
	switch {
	case (fieldType == "string" || fieldType == "null.String") && operator == "=":
		return fmt.Sprintf("%s ILIKE ?", fieldName), "%" + fmt.Sprint(value) + "%"
	case fieldType == "map[string]interface {}": // JSONB
		jsonValue, _ := json.Marshal(value)
		return fmt.Sprintf("%s @> ?", fieldName), string(jsonValue)
	default:
		return fmt.Sprintf("%s %s ?", fieldName, operator), value
	}
}

func addFilter2[T any](query map[string]interface{}) (string, []interface{}) {
	allowedFields := extractAllowedFields[T]()
	var filter strings.Builder
	var params []interface{}
	filter.WriteString("1=1")

	for rawKey, value := range query {
		fieldName, operator := parseKey(rawKey)
		if fieldType, exists := allowedFields[fieldName]; exists {
			condition, param := buildSql(fieldName, fieldType, operator, value)
			filter.WriteString(" AND " + condition)
			params = append(params, param)
		}
	}
	return filter.String(), params
	// return func(db *gorm.DB) *gorm.DB {
	// 	return db.Where(gorm.Expr(filter.String(), params...))
	// }
}
