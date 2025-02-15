package utils

import "regexp"

func IsPhoneNumber(phone string) bool {
	// Check if the phone number is valid
	phoneValidator := regexp.MustCompile(`^(\+?(\d{1,3}))?(\d{10,15})$`)
	return phoneValidator.MatchString(phone)
}