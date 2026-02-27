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
		return &Error{Field: "name", Message: "Full name cannot be empty"}
	}

	if len(name) < minNameLength {
		return &Error{Field: "name", Message: fmt.Sprintf("The full name must contain at least a %d character", minNameLength)}
	}

	if len(name) > maxNameLength {
		return &Error{Field: "name", Message: fmt.Sprintf("The full name must contain a maximum of %d characters", maxNameLength)}
	}

	// We allow letters (Russian and English), spaces and hyphens for double surnames
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ\s-]+$`, name)
	if !match {
		return &Error{Field: "name", Message: "The full name can contain only letters, spaces, and hyphens"}
	}

	if strings.Contains(name, "  ") {
		return &Error{Field: "name", Message: "Full name should not contain double spaces"}
	}

	if len(strings.Fields(name)) < 2 {
		return &Error{Field: "name", Message: "Enter your last name and first name"}
	}

	return nil
}

// Organization validates organization name input
func Organization(org string) error {
	org = strings.TrimSpace(org)

	if org == "" {
		return &Error{Field: "organization", Message: "The name of the organization cannot be empty"}
	}

	if len(org) < minOrgLength {
		return &Error{Field: "organization", Message: fmt.Sprintf("The name of the organization must contain at least %d characters", minOrgLength)}
	}

	if len(org) > maxOrgLength {
		return &Error{Field: "organization", Message: fmt.Sprintf("The name of the organization must contain a maximum of %d characters", maxOrgLength)}
	}

	// We allow letters, numbers, spaces, quotation marks, dots, commas, hyphens, ampersands, numbers and slashes
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ0-9\s"'\.,\-&№/]+$`, org)
	if !match {
		return &Error{Field: "organization", Message: "The organization's name contains invalid characters"}
	}

	return nil
}

// Position validates job position input
func Position(position string) error {
	position = strings.TrimSpace(position)

	if position == "" {
		return &Error{Field: "position", Message: "The position cannot be empty"}
	}

	if len(position) < minPositionLength {
		return &Error{Field: "position", Message: fmt.Sprintf("The post must contain at least a %d character", minPositionLength)}
	}

	if len(position) > maxPositionLength {
		return &Error{Field: "position", Message: fmt.Sprintf("The post must contain a maximum of %d characters", maxPositionLength)}
	}

	// We allow letters, numbers, spaces, hyphens, dots, commas, slashes
	match, _ := regexp.MatchString(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-.,/]+$`, position)
	if !match {
		return &Error{Field: "position", Message: "The post contains invalid characters"}
	}

	return nil
}
