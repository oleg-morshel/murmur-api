package auth_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
	"github.com/oleg-morshel/murmur-api/internal/testutil"
)

func registerBody(username, email, password string) string {
	return fmt.Sprintf(`{"username":%q,"email":%q,"password":%q}`, username, email, password)
}

func TestAuthHTTPHandler_Register(t *testing.T) {
	const (
		username = "oleg"
		email    = "oleg@example.com"
		password = "super-secret"
	)

	tokens := &auth_service.TokenPair{AccessToken: "access-123", RefreshToken: "refresh-456"}

	tests := []struct {
		name              string
		body              string
		svcTokens         *auth_service.TokenPair
		svcErr            error
		wantStatus        int
		wantSvcCalls      int
		wantMessage       string
		wantErrorContains string
	}{
		{
			name:         "success",
			body:         registerBody(username, email, password),
			svcTokens:    tokens,
			wantStatus:   http.StatusCreated,
			wantSvcCalls: 1,
		},
		{
			name:              "duplicate email",
			body:              registerBody(username, email, password),
			svcErr:            fmt.Errorf("auth.Register: %w", core_errors.ErrConflict),
			wantStatus:        http.StatusConflict,
			wantSvcCalls:      1,
			wantMessage:       "user already exists",
			wantErrorContains: "conflict",
		},
		{
			name:              "malformed json",
			body:              `{not json`,
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "invalid request body",
		},
		{
			name:              "empty body",
			body:              ``,
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "invalid request body",
		},
		{
			name:              "unknown field is rejected",
			body:              `{"username":"oleg","email":"oleg@example.com","password":"super-secret","admin":true}`,
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "invalid request body",
		},
		{
			name:              "username too short",
			body:              registerBody("ab", email, password),
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "username must be at least 3 characters",
		},
		{
			name:              "username not alphanumeric",
			body:              registerBody("ol eg!", email, password),
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "username must contain only letters and numbers",
		},
		{
			name:              "invalid email",
			body:              registerBody(username, "not-an-email", password),
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "email must be a valid email",
		},
		{
			name:              "password too short",
			body:              registerBody(username, email, "short"),
			wantStatus:        http.StatusBadRequest,
			wantMessage:       "invalid request body",
			wantErrorContains: "password must be at least 8 characters",
		},
		{
			name:              "unexpected service error",
			body:              registerBody(username, email, password),
			svcErr:            errors.New("db down"),
			wantStatus:        http.StatusInternalServerError,
			wantSvcCalls:      1,
			wantMessage:       "internal error",
			wantErrorContains: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svcCalls                         int
				gotUsername, gotEmail, gotPasswd string
			)

			svc := &mockAuthService{
				registerFn: func(_ context.Context, u, e, p string) (*auth_service.TokenPair, error) {
					svcCalls++
					gotUsername, gotEmail, gotPasswd = u, e, p
					return tt.svcTokens, tt.svcErr
				},
			}
			handler := NewAuthHTTPHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body)).
				WithContext(testutil.NewContext())
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if svcCalls != tt.wantSvcCalls {
				t.Errorf("service calls = %d, want %d", svcCalls, tt.wantSvcCalls)
			}

			if tt.wantStatus == http.StatusCreated {
				var got auth_service.TokenPair
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("response is not valid JSON: %v", err)
				}
				if got != *tt.svcTokens {
					t.Errorf("tokens = %+v, want %+v", got, *tt.svcTokens)
				}
				if gotUsername != username || gotEmail != email || gotPasswd != password {
					t.Errorf("service got (%q, %q, %q), want (%q, %q, %q)",
						gotUsername, gotEmail, gotPasswd, username, email, password)
				}
				return
			}

			var resp map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("error response is not valid JSON: %v", err)
			}
			if resp["message"] != tt.wantMessage {
				t.Errorf("message = %q, want %q", resp["message"], tt.wantMessage)
			}
			if !strings.Contains(resp["error"], tt.wantErrorContains) {
				t.Errorf("error = %q, want it to contain %q", resp["error"], tt.wantErrorContains)
			}
		})
	}
}
