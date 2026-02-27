package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid full name",
			input:    "Иванов Иван Иванович",
			expected: true,
		},
		{
			name:     "Valid with extra spaces",
			input:    "  Петров Петр Петрович  ",
			expected: true,
		},
		{
			name:     "Valid two words",
			input:    "Сидоров Сидор",
			expected: true,
		},
		{
			name:     "Valid English",
			input:    "John Doe Smith",
			expected: true,
		},
		{
			name:     "Valid with hyphen",
			input:    "Салтыков-Щедрин Михаил",
			expected: true,
		},
		{
			name:     "Valid mixed",
			input:    "Анна Maria Schmidt",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Name(tc.input)
			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestName_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{
			name:        "Empty name",
			input:       "",
			expectedMsg: "ФИО не может быть пустым",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedMsg: "ФИО не может быть пустым",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedMsg: "ФИО должно содержать минимум 3 символа",
		},
		{
			name:        "Too long",
			input:       strings.Repeat("A", 101),
			expectedMsg: "ФИО должно содержать максимум 100 символов",
		},
		{
			name:        "Contains numbers",
			input:       "Иванов Иван 123",
			expectedMsg: "ФИО может содержать только буквы, пробелы и дефисы",
		},
		{
			name:        "Contains special chars",
			input:       "Иванов Иван!",
			expectedMsg: "ФИО может содержать только буквы, пробелы и дефисы",
		},
		{
			name:        "Double spaces",
			input:       "Иванов  Иван",
			expectedMsg: "ФИО не должно содержать двойных пробелов",
		},
		{
			name:        "Single word",
			input:       "Иванов",
			expectedMsg: "Укажите фамилию и имя",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Name(tc.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)

			// Проверяем, что это наш тип ошибки
			valErr, ok := err.(*Error)
			assert.True(t, ok)
			assert.Equal(t, "name", valErr.Field)
		})
	}
}

func TestOrganization_Valid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple name",
			input: "ООО Ромашка",
		},
		{
			name:  "With quotes",
			input: `ООО "Ромашка"`,
		},
		{
			name:  "With hyphen",
			input: "ООО Ромашка-Флора",
		},
		{
			name:  "With numbers",
			input: "Компания №1",
		},
		{
			name:  "With ampersand",
			input: "Johnson & Johnson",
		},
		{
			name:  "With dot",
			input: "ООО Ромашка и Ко.",
		},
		{
			name:  "With slash",
			input: "ООО Ромашка/Флора",
		},
		{
			name:  "Complex name",
			input: `ООО "Ромашка-Флора" №1`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Organization(tc.input)
			assert.NoError(t, err)
		})
	}
}

func TestOrganization_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{
			name:        "Empty",
			input:       "",
			expectedMsg: "Название организации не может быть пустым",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedMsg: "должно содержать минимум 2 символа",
		},
		{
			name:        "Invalid char @",
			input:       "Company@Name",
			expectedMsg: "недопустимые символы",
		},
		{
			name:        "Invalid char #",
			input:       "Company#Name",
			expectedMsg: "недопустимые символы",
		},
		{
			name:        "Invalid char $",
			input:       "Company$Name",
			expectedMsg: "недопустимые символы",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Organization(tc.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)

			valErr, ok := err.(*Error)
			assert.True(t, ok)
			assert.Equal(t, "organization", valErr.Field)
		})
	}
}

func TestPosition_Valid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple position",
			input: "Менеджер",
		},
		{
			name:  "Compound position",
			input: "Старший разработчик Go",
		},
		{
			name:  "With hyphen",
			input: "Бизнес-аналитик",
		},
		{
			name:  "With slash",
			input: "DevOps/SRE инженер",
		},
		{
			name:  "English",
			input: "Senior Software Engineer",
		},
		{
			name:  "With numbers",
			input: "Разработчик 1С",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Position(tc.input)
			assert.NoError(t, err)
		})
	}
}

func TestPosition_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{
			name:        "Empty",
			input:       "",
			expectedMsg: "Должность не может быть пустой",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedMsg: "должна содержать минимум 2 символа",
		},
		{
			name:        "Invalid char @",
			input:       "Manager@Company",
			expectedMsg: "недопустимые символы",
		},
		{
			name:        "Invalid char #",
			input:       "Manager#2",
			expectedMsg: "недопустимые символы",
		},
		{
			name:        "Invalid char $",
			input:       "Manager$",
			expectedMsg: "недопустимые символы",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Position(tc.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)

			valErr, ok := err.(*Error)
			assert.True(t, ok)
			assert.Equal(t, "position", valErr.Field)
		})
	}
}
