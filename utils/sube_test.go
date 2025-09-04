package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

// CheckSubeValid checks if the sube is a valid JSON array of UUIDs

func TestSubeValid(t *testing.T) {
	subeValid := datatypes.JSON(`["123e4567-e89b-12d3-a456-426614174000","123e4567-e89b-12d3-a456-426614174001"]`)
	subeInvalid := datatypes.JSON(`["invalid-uuid","123e4567-e89b-12d3-a456-426614174001"]`)
	subeEmpty := datatypes.JSON(`[]`)
	subeNil := datatypes.JSON(nil)
	subeWithoutBrackets := datatypes.JSON(`"123e4567-e89b-12d3-a456-426614174000"`)

	assert := assert.New(t)
	// Test valid sube
	valid, err := CheckSubeValid(subeValid)
	assert.Equal(true, valid)
	assert.Nil(err)

	// Test invalid sube
	valid, err = CheckSubeValid(subeInvalid)
	assert.Equal(false, valid)
	assert.NotNil(err)

	// Test empty sube
	valid, err = CheckSubeValid(subeEmpty)
	assert.Equal(false, valid)
	assert.NotNil(err)

	// Test nil sube
	valid, err = CheckSubeValid(subeNil)
	assert.Equal(false, valid)
	assert.NotNil(err)

	// Test sube without brackets
	valid, err = CheckSubeValid(subeWithoutBrackets)
	assert.Equal(false, valid)
	assert.NotNil(err)
}
