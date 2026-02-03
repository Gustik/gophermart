package zaplog

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"Valid level", "info", false},
		{"Invalid level", "unexisting_level", true},
		{"Empty level", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() ошибка = %v, ожидалось наличие ошибки: %v", err, tt.wantErr)
				return
			}
			if err == nil && logger == nil {
				t.Error("New() вернул nil вместо логгера, хотя ошибка отсутствует")
			}
		})
	}
}
