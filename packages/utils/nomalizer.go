package utils

import "reflect"

func NormalizeFields(payload interface{}) {
	val := reflect.ValueOf(payload).Elem() // Dereference pointer to get the struct

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Only process string fields
		if field.Kind() == reflect.String {
			// Check if the field is "string" and set it to ""
			if field.String() == "string" {
				field.SetString("")
			}
		}

		// Handle nested structs (if any)
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			NormalizeFields(field.Interface())
		}
	}
}
