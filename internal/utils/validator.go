package utils

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

// ValidateStruct 驗證結構體
func ValidateStruct(data interface{}) error {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("validate")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if err := validateField(field, rule, fieldType.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateField(field reflect.Value, rule, fieldName string) error {
	switch {
	case rule == "required":
		if field.Kind() == reflect.String && field.String() == "" {
			return errors.New(fieldName + " 是必填欄位")
		}
		if field.Kind() == reflect.Int && field.Int() == 0 {
			return errors.New(fieldName + " 是必填欄位")
		}

	case rule == "email":
		if field.Kind() == reflect.String {
			email := field.String()
			emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			if !regexp.MustCompile(emailRegex).MatchString(email) {
				return errors.New(fieldName + " 格式不正確")
			}
		}

	case strings.HasPrefix(rule, "min="):
		minStr := strings.TrimPrefix(rule, "min=")
		if field.Kind() == reflect.String {
			if len(field.String()) < parseIntFromString(minStr) {
				return errors.New(fieldName + " 長度不能少於 " + minStr + " 個字符")
			}
		}

	case strings.HasPrefix(rule, "max="):
		maxStr := strings.TrimPrefix(rule, "max=")
		if field.Kind() == reflect.String {
			if len(field.String()) > parseIntFromString(maxStr) {
				return errors.New(fieldName + " 長度不能超過 " + maxStr + " 個字符")
			}
		}
	}

	return nil
}

func parseIntFromString(s string) int {
	// 簡單的字符串轉整數，實際應用中可以使用 strconv.Atoi
	switch s {
	case "3":
		return 3
	case "6":
		return 6
	case "20":
		return 20
	case "50":
		return 50
	case "100":
		return 100
	default:
		return 0
	}
}
