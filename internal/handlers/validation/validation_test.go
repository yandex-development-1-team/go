package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Standard Russian name",
			input: "Иванов Иван Иванович",
		},
		{
			name:  "Standard English name",
			input: "John Doe",
		},
		{
			name:  "Name with hyphen",
			input: "Салтыков-Щедрин Михаил",
		},
		{
			name:  "Name with single spaces only",
			input: "Петров Петр Петрович",
		},
		{
			name:  "Name with leading/trailing spaces (should be trimmed)",
			input: "  Сидоров Сидор Сидорович  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Name(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestName_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Empty name",
			input:       "",
			expectedErr: "Full name cannot be empty",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "Full name cannot be empty",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "The full name must contain at least a 3 character",
		},
		{
			name:        "Too long",
			input:       "This is a very long name that exceeds the maximum allowed length of one hundred characters which should be more than enough for any normal person",
			expectedErr: "The full name must contain a maximum of 100 characters",
		},
		{
			name:        "Contains numbers",
			input:       "Ivan123 Petrov",
			expectedErr: "The full name can contain only letters, spaces, and hyphens",
		},
		{
			name:        "Contains special chars",
			input:       "Ivan@ Petrov",
			expectedErr: "The full name can contain only letters, spaces, and hyphens",
		},
		{
			name:        "Double spaces",
			input:       "Ivan  Petrov",
			expectedErr: "Full name should not contain double spaces",
		},
		{
			name:        "Multiple double spaces",
			input:       "Ivan  Petrov  Ivanovich",
			expectedErr: "Full name should not contain double spaces",
		},
		{
			name:        "Single word",
			input:       "Ivan",
			expectedErr: "Enter your last name and first name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Name(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestOrganization_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Standard organization",
			input: "ООО Ромашка",
		},
		{
			name:  "Organization with quotes",
			input: "ООО \"Ромашка\"",
		},
		{
			name:  "Organization with number",
			input: "Компания №1",
		},
		{
			name:  "Organization with hyphen",
			input: "Рога-и-Копыта",
		},
		{
			name:  "Organization with ampersand",
			input: "Johnson & Johnson",
		},
		{
			name:  "Organization with leading/trailing spaces",
			input: "  ООО Ромашка  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Organization(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestOrganization_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Empty organization",
			input:       "",
			expectedErr: "The name of the organization cannot be empty",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "The name of the organization cannot be empty",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "The name of the organization must contain at least 2 characters",
		},
		{
			name:        "Too long",
			input:       "This is a very long organization name that exceeds the maximum allowed length of two hundred fifty five characters which should be more than enough for any normal company name in the world even with all the legal suffixes and prefixes that might be required by law in some jurisdictions",
			expectedErr: "The name of the organization must contain a maximum of 255 characters",
		},
		{
			name:        "Contains invalid chars",
			input:       "Company@#$%",
			expectedErr: "The organization's name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Organization(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestPosition_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Standard position",
			input: "Менеджер по продажам",
		},
		{
			name:  "Position with hyphen",
			input: "IT-директор",
		},
		{
			name:  "Position with dot",
			input: "Ph.D. Researcher",
		},
		{
			name:  "Position with slash",
			input: "UI/UX Designer",
		},
		{
			name:  "Position with number",
			input: "Разработчик 1С",
		},
		{
			name:  "Position with leading/trailing spaces",
			input: "  Старший инженер  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Position(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestPosition_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Empty position",
			input:       "",
			expectedErr: "The position cannot be empty",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "The position cannot be empty",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "The post must contain at least a 2 character",
		},
		{
			name:        "Too long",
			input:       "This is a very long position title that exceeds the maximum allowed length of one hundred characters which is really too long for any job title in the world",
			expectedErr: "The post must contain a maximum of 100 characters",
		},
		{
			name:        "Contains invalid chars",
			input:       "Engineer@#$%",
			expectedErr: "The post contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Position(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
