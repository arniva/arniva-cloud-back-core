package utils

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"
)

func CheckSubeValid(sube datatypes.JSON) (bool, error) {
	if sube == nil || len(sube) == 0 {
		return false, errors.New("sube nil veya bos olamaz")
	}

	var subeList []string
	if err := json.Unmarshal(sube, &subeList); err != nil {
		return false, errors.New("sube JSON array formatinda olmali")
	}
	if len(subeList) == 0 {
		return false, errors.New("sube bos olamaz")
	}

	for _, subeItem := range subeList {
		subeItem = strings.TrimSpace(subeItem)
		if !isValidUUID(subeItem) {
			return false, errors.New("sube geçersiz UUID formatinda")
		}
	}

	return true, nil
}

func isValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}
	parts := strings.Split(u, "-")
	if len(parts) != 5 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}

	for _, c := range u {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') || c == '-') {
			return false
		}
	}
	return true
}
