package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	helloHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	expected := "Hello, World!"
	if body != expected {
		t.Errorf("expected body %q, got %q", expected, body)
	}
}

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{"add", "a=3&b=2&op=add", 200, "5"},
		{"sub", "a=10&b=4&op=sub", 200, "6"},
		{"mul", "a=3&b=5&op=mul", 200, "15"},
		{"div", "a=10&b=4&op=div", 200, "2.5"},
		{"div by zero", "a=1&b=0&op=div", 400, ""},
		{"invalid op", "a=1&b=2&op=pow", 400, ""},
		{"missing a", "b=2&op=add", 400, ""},
		{"missing b", "a=2&op=add", 400, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/calc?"+tt.query, nil)
			w := httptest.NewRecorder()

			calcHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantStatus == 200 && w.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}
