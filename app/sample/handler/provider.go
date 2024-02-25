package handler

import (
	"context"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"strings"
)

type Provider struct {
	d Dependencies
}

type Dependencies struct {
	Kratos *kratos.Provider
}

type NewInput struct{}

func New(i NewInput, d Dependencies) (*Provider, error) {
	p := Provider{
		d: d,
	}
	return &p, nil
}

func (p *Provider) RegisterHandles(mux *http.ServeMux) *http.ServeMux {
	// Static files
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("GET /static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			fileServer.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))

	// health check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Authentication Registration
	mux.Handle("GET /auth/registration", p.baseMiddleware(p.handleGetAuthRegistration))
	mux.Handle("POST /auth/registration", p.baseMiddleware(p.handlePostAuthRegistration))
	mux.Handle("POST /auth/registration/oidc", p.baseMiddleware(p.handlePostAuthRegistrationOidc))

	// Authentication Verification
	mux.Handle("GET /auth/verification", p.baseMiddleware(p.handleGetAuthVerification))
	mux.Handle("GET /auth/verification/code", p.baseMiddleware(p.handleGetAuthVerificationCode))
	mux.Handle("POST /auth/verification/email", p.baseMiddleware(p.handlePostVerificationEmail))
	mux.Handle("POST /auth/verification/code", p.baseMiddleware(p.handlePostVerificationCode))

	// Authentication Login
	mux.Handle("GET /auth/login", p.baseMiddleware(p.handleGetAuthLogin))
	mux.Handle("POST /auth/login", p.baseMiddleware(p.handlePostAuthLogin))
	mux.Handle("POST /auth/login/oidc", p.baseMiddleware(p.handlePostAuthLoginOidc))

	// Authentication Logout
	mux.Handle("POST /auth/logout", p.baseMiddleware(p.handlePostAuthLogout))

	// Authentication Recovery
	mux.Handle("GET /auth/recovery", p.baseMiddleware(p.handleGetAuthRecovery))
	mux.Handle("POST /auth/recovery/email", p.baseMiddleware(p.handlePostAuthRecoveryEmail))
	mux.Handle("POST /auth/recovery/code", p.baseMiddleware(p.handlePostAuthRecoveryCode))

	// My Password
	mux.Handle("GET /my/password", p.baseMiddleware(p.handleGetMyPassword))
	mux.Handle("POST /my/password", p.baseMiddleware(p.handlePostMyPassword))

	// My Profile
	mux.Handle("GET /my/profile", p.baseMiddleware(p.handleGetMyProfile))
	mux.Handle("GET /my/profile/edit", p.baseMiddleware(p.handleGetMyProfileEdit))
	mux.Handle("GET /my/profile/form", p.baseMiddleware(p.handleGetMyProfileForm))
	mux.Handle("POST /my/profile", p.baseMiddleware(p.handlePostMyProfile))

	// Top
	mux.Handle("GET /", p.baseMiddleware(p.handleGetTop))

	// Item
	mux.Handle("GET /item/{id}", p.baseMiddleware(p.handleGetItemDetail))
	mux.Handle("GET /item/{id}/purchase", p.baseMiddleware(p.handleGetItemPurchase))
	mux.Handle("POST /item/{id}/purchase", p.baseMiddleware(p.handlePostItemPurchase))

	return mux
}

func (p *Provider) baseMiddleware(handler http.HandlerFunc) http.Handler {
	return p.loggingRquest(
		p.setSession(handler),
	)
}

func (p *Provider) loggingRquest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.Info(fmt.Sprintf("[Request] %s %s", r.Method, r.URL.Path))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (p *Provider) setSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		output, err := p.d.Kratos.ToSession(kratos.ToSessionInput{
			Cookie: r.Header.Get("Cookie"),
		})
		if err != nil {
			ctx = context.WithValue(ctx, "session", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, "session", output.Session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
