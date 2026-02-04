package validator

import (
	"errors"
	"strconv"
)

var (
	ErrEmptyOrderNumber   = errors.New("номер заказа не может быть пустым")
	ErrInvalidOrderFormat = errors.New("неверный формат номера заказа")
)

// ValidateOrderNumber проверяет номер заказа
func ValidateOrderNumber(number string) error {
	if number == "" {
		return ErrEmptyOrderNumber
	}

	if !ValidateLuhn(number) {
		return ErrInvalidOrderFormat
	}

	return nil
}

// ValidateLuhn проверяет номер по алгоритму Луна
func ValidateLuhn(number string) bool {
	// Пустая строка невалидна
	if len(number) == 0 {
		return false
	}

	// Проверяем, что все символы - цифры
	for _, ch := range number {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	sum := 0
	parity := len(number) % 2

	for i, digit := range number {
		// Преобразуем символ в число
		n, _ := strconv.Atoi(string(digit))

		// Удваиваем каждую вторую цифру (с конца)
		if i%2 == parity {
			n *= 2
			// Если результат > 9, вычитаем 9
			if n > 9 {
				n -= 9
			}
		}

		sum += n
	}

	// Номер валиден, если сумма кратна 10
	return sum%10 == 0
}
