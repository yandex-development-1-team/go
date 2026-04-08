package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Configuration constants
const (
	minNameLength     = 3
	maxNameLength     = 100
	minOrgLength      = 2
	maxOrgLength      = 255
	minPositionLength = 2
	maxPositionLength = 100
)

// Error represents a validation error with user-friendly message
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// Name checks the name
func Name(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return &Error{Field: "name", Message: "Полное имя не может быть пустым"}
	}

	if len(name) < minNameLength {
		return &Error{Field: "name", Message: fmt.Sprintf("Полное имя должно содержать как минимум %d символа", minNameLength)}
	}

	if len(name) > maxNameLength {
		return &Error{Field: "name", Message: fmt.Sprintf("Полное имя должно содержать не более %d символов", maxNameLength)}
	}

	// We allow letters (Russian and English), spaces and hyphens for double surnames
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ\s-]+$`, name)
	if !match {
		return &Error{Field: "name", Message: "Полное имя может содержать только буквы, пробелы и дефисы"}
	}

	if strings.Contains(name, "  ") {
		return &Error{Field: "name", Message: "Полное имя не должно содержать двойных пробелов"}
	}

	if len(strings.Fields(name)) < 2 {
		return &Error{Field: "name", Message: "Введите вашу фамилию и имя"}
	}

	return nil
}

// Organization validates organization name input
func Organization(org string) error {
	org = strings.TrimSpace(org)

	if org == "" {
		return &Error{Field: "organization", Message: "Название организации не может быть пустым"}
	}

	if len(org) < minOrgLength {
		return &Error{Field: "organization", Message: fmt.Sprintf("Название организации должно содержать не менее %d символов", minOrgLength)}
	}

	if len(org) > maxOrgLength {
		return &Error{Field: "organization", Message: fmt.Sprintf("Название организации не должно содержать более %d символов", maxOrgLength)}
	}

	// We allow letters, numbers, spaces, quotation marks, dots, commas, hyphens, ampersands, numbers and slashes
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ0-9\s"'\.,\-&№/]+$`, org)
	if !match {
		return &Error{Field: "organization", Message: "Название организации содержит недопустимые символы"}
	}

	return nil
}

// Position validates job position input
func Position(position string) error {
	position = strings.TrimSpace(position)

	if position == "" {
		return &Error{Field: "position", Message: "Должность не может быть пустой"}
	}

	if len(position) < minPositionLength {
		return &Error{Field: "position", Message: fmt.Sprintf("Сообщение должно содержать как минимум %d символа", minPositionLength)}
	}

	if len(position) > maxPositionLength {
		return &Error{Field: "position", Message: fmt.Sprintf("Сообщение должно содержать максимум %d символов", maxPositionLength)}
	}

	// We allow letters, numbers, spaces, hyphens, dots, commas, slashes
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-.,/]+$`, position)
	if !match {
		return &Error{Field: "position", Message: "В сообщении содержатся недопустимые символы"}
	}

	return nil
}
