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
			expectedErr: "Полное имя не может быть пустым",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "Полное имя не может быть пустым",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "Полное имя должно содержать как минимум 3 символа",
		},
		{
			name:        "Too long",
			input:       "This is a very long name that exceeds the maximum allowed length of one hundred characters which should be more than enough for any normal person",
			expectedErr: "Полное имя должно содержать не более 100 символов",
		},
		{
			name:        "Contains numbers",
			input:       "Ivan123 Petrov",
			expectedErr: "Полное имя может содержать только буквы, пробелы и дефисы",
		},
		{
			name:        "Contains special chars",
			input:       "Ivan@ Petrov",
			expectedErr: "Полное имя может содержать только буквы, пробелы и дефисы",
		},
		{
			name:        "Double spaces",
			input:       "Ivan  Petrov",
			expectedErr: "Полное имя не должно содержать двойных пробелов",
		},
		{
			name:        "Multiple double spaces",
			input:       "Ivan  Petrov  Ivanovich",
			expectedErr: "Полное имя не должно содержать двойных пробелов",
		},
		{
			name:        "Single word",
			input:       "Ivan",
			expectedErr: "Введите вашу фамилию и имя",
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
			expectedErr: "Название организации не может быть пустым",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "Название организации не может быть пустым",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "Название организации должно содержать не менее 2 символов",
		},
		{
			name:        "Too long",
			input:       "This is a very long organization name that exceeds the maximum allowed length of two hundred fifty five characters which should be more than enough for any normal company name in the world even with all the legal suffixes and prefixes that might be required by law in some jurisdictions",
			expectedErr: "Название организации не должно содержать более 255 символов",
		},
		{
			name:        "Contains invalid chars",
			input:       "Company@#$%",
			expectedErr: "Название организации содержит недопустимые символы",
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
			expectedErr: "Должность не может быть пустой",
		},
		{
			name:        "Only spaces",
			input:       "   ",
			expectedErr: "Должность не может быть пустой",
		},
		{
			name:        "Too short",
			input:       "A",
			expectedErr: "Сообщение должно содержать как минимум 2 символа",
		},
		{
			name:        "Too long",
			input:       "This is a very long position title that exceeds the maximum allowed length of one hundred characters which is really too long for any job title in the world",
			expectedErr: "Сообщение должно содержать максимум 100 символов",
		},
		{
			name:        "Contains invalid chars",
			input:       "Engineer@#$%",
			expectedErr: "В сообщении содержатся недопустимые символы",
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
