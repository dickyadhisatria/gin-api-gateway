package usecase_test

import (
	"errors"
	"gin-api-gateway/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockProxyClient struct {
	serveFunc func(rw http.ResponseWriter, req *http.Request, target string) error
}

func (m *mockProxyClient) Serve(rw http.ResponseWriter, req *http.Request, target string) error {
	return m.serveFunc(rw, req, target)
}

func TestProxyUseCase_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		expectedTarget := "http://localhost:3001"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
		rec := httptest.NewRecorder()

		client := &mockProxyClient{
			serveFunc: func(rw http.ResponseWriter, r *http.Request, target string) error {
				if target != expectedTarget {
					t.Errorf("expected target %s, got %s", expectedTarget, target)
				}
				rw.WriteHeader(http.StatusOK)
				return nil
			},
		}

		uc := usecase.NewProxyUseCase(client)
		err := uc.Execute(rec, req, expectedTarget)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("downstream error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
		rec := httptest.NewRecorder()
		expectedErr := errors.New("downstream failure")

		client := &mockProxyClient{
			serveFunc: func(rw http.ResponseWriter, r *http.Request, target string) error {
				return expectedErr
			},
		}

		uc := usecase.NewProxyUseCase(client)
		err := uc.Execute(rec, req, "http://localhost:3001")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}
