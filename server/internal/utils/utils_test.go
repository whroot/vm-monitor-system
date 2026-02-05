package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStringPtr(t *testing.T) {
	input := "test string"
	result := StringPtr(input)

	assert.NotNil(t, result)
	assert.Equal(t, input, *result)
}

func TestTimePtr(t *testing.T) {
	input := time.Now()
	result := TimePtr(input)

	assert.NotNil(t, result)
	assert.Equal(t, input.Unix(), result.Unix())
}

func TestHashString(t *testing.T) {
	t.Run("HashConsistency", func(t *testing.T) {
		input := "password123"
		hash1 := HashString(input)
		hash2 := HashString(input)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("DifferentInputs", func(t *testing.T) {
		hash1 := HashString("password1")
		hash2 := HashString("password2")

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("EmptyString", func(t *testing.T) {
		hash := HashString("")
		assert.NotEmpty(t, hash)
	})

	t.Run("HashLength", func(t *testing.T) {
		hash := HashString("test")
		assert.Len(t, hash, 64)
	})
}

func TestValidatePasswordComplexity(t *testing.T) {
	t.Run("ValidPassword", func(t *testing.T) {
		password := "Test@123"
		result := ValidatePasswordComplexity(password)
		assert.True(t, result)
	})

	t.Run("TooShort", func(t *testing.T) {
		password := "T@1"
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("MissingUppercase", func(t *testing.T) {
		password := "test@123"
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("MissingLowercase", func(t *testing.T) {
		password := "TEST@123"
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("MissingNumber", func(t *testing.T) {
		password := "Test@abc"
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("MissingSpecialChar", func(t *testing.T) {
		password := "Test123"
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		password := ""
		result := ValidatePasswordComplexity(password)
		assert.False(t, result)
	})

	t.Run("ComplexPassword", func(t *testing.T) {
		password := "MyP@ssw0rd!2024"
		result := ValidatePasswordComplexity(password)
		assert.True(t, result)
	})

	t.Run("EdgeCaseMinLength", func(t *testing.T) {
		password := "Ab1@abcd"
		result := ValidatePasswordComplexity(password)
		assert.True(t, result)
	})
}
