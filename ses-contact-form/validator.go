package validator

import (
    "fmt"
    "regexp"
)

var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var PhoneRegex = regexp.MustCompile(`^\+?[\d\s\-()]{10,}$`)

func ValidateEmail(email string) error {
    if !EmailRegex.MatchString(email) {
        return fmt.Errorf("Invalid email format")
    }

    return nil
}

func ValidatePhone(phone string) error {
    if !PhoneRegex.MatchString(phone) {
        return fmt.Errorf("Invalid phone format")
    }

    return nil
}

func SanitizeInput(input string) string {
    // Remove sus characters
    sanitized := regexp.MustCompile(`[<>]`).ReplaceAllString(input, "")
    return sanitized
}

func ValidateRequired(fieldName, value string) error {
    if value == "" {
        return fmt.Errorf("Field '%s' is required", fieldName)
    }

    return nil
}
