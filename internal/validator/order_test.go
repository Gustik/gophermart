package validator

import "testing"

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "Валидный номер - 79927398713",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "Валидный номер - 12345678903",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "Валидный номер - 4561261212345467",
			number: "4561261212345467",
			want:   true,
		},
		{
			name:   "Валидный номер - 0",
			number: "0",
			want:   true,
		},
		{
			name:   "Невалидный номер - 79927398710",
			number: "79927398710",
			want:   false,
		},
		{
			name:   "Невалидный номер - 12345678901",
			number: "12345678901",
			want:   false,
		},
		{
			name:   "Невалидный номер - содержит буквы",
			number: "1234567890a",
			want:   false,
		},
		{
			name:   "Невалидный номер - содержит пробелы",
			number: "1234 5678 903",
			want:   false,
		},
		{
			name:   "Невалидный номер - содержит спецсимволы",
			number: "1234-5678-903",
			want:   false,
		},
		{
			name:   "Невалидный номер - пустая строка",
			number: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateLuhn(tt.number)
			if got != tt.want {
				t.Errorf("ValidateLuhn(%q) = %v, ожидалось %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr error
	}{
		{
			name:    "Валидный номер заказа",
			number:  "79927398713",
			wantErr: nil,
		},
		{
			name:    "Валидный номер заказа - короткий",
			number:  "0",
			wantErr: nil,
		},
		{
			name:    "Пустой номер заказа",
			number:  "",
			wantErr: ErrEmptyOrderNumber,
		},
		{
			name:    "Неверный формат - неправильная контрольная сумма",
			number:  "79927398710",
			wantErr: ErrInvalidOrderFormat,
		},
		{
			name:    "Неверный формат - содержит буквы",
			number:  "7992739871a",
			wantErr: ErrInvalidOrderFormat,
		},
		{
			name:    "Неверный формат - содержит пробелы",
			number:  "7992 739 8713",
			wantErr: ErrInvalidOrderFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderNumber(tt.number)
			if err != tt.wantErr {
				t.Errorf("ValidateOrderNumber(%q) ошибка = %v, ожидалась %v", tt.number, err, tt.wantErr)
			}
		})
	}
}
