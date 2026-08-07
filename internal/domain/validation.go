package domain

import "unicode"

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func IsValidLuhn(number string) bool {
	if !isDigits(number) {
		return false
	}

	sum := 0
	parity := len(number) % 2
	for i, r := range number {
		digit := int(r - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
