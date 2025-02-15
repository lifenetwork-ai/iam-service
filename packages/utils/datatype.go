package utils

import "regexp"

func IsPhoneNumber(phone string) bool {
	// Check if the phone number is valid
	phoneValidator := regexp.MustCompile(`^(\+?(\d{1,3}))?(\d{10,15})$`)
	return phoneValidator.MatchString(phone)
}

func IsEmail(email string) bool {
	// Check if the email is valid
	emailValidator := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailValidator.MatchString(email)
}
