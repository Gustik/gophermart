package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Переменная окружения установлена",
			key:          "TEST_VAR_1",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
		},
		{
			name:         "Переменная окружения не установлена",
			key:          "TEST_VAR_2",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "Пустая переменная окружения",
			key:          "TEST_VAR_3",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменную окружения, если нужно
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				// Убеждаемся, что переменная не установлена
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, ожидалось %v", result, tt.expected)
			}
		})
	}
}
