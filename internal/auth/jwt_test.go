package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() http.Request
		wantUserID int
		wantOk     bool
	}{
		{
			name: "UserID присутствует в контексте",
			setupCtx: func() http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Используем auth.SetUserID, а проверяем auth.GetUserID
				ctx := SetUserID(req.Context(), 789)
				return *req.WithContext(ctx)
			},
			wantUserID: 789,
			wantOk:     true,
		},
		{
			name: "UserID отсутствует в контексте",
			setupCtx: func() http.Request {
				return *httptest.NewRequest(http.MethodGet, "/", nil)
			},
			wantUserID: 0,
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupCtx()
			gotUserID, gotOk := GetUserID(req.Context())

			if gotOk != tt.wantOk {
				t.Errorf("GetUserID() ok = %v, ожидалось %v", gotOk, tt.wantOk)
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("GetUserID() userID = %v, ожидалось %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
