package handler

import (
	"context"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"slices"
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
	mux.Handle(fmt.Sprintf("GET %s", routePaths.AuthRegistration), p.baseMiddleware(p.handleGetAuthRegistration))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthRegistration), p.baseMiddleware(p.handlePostAuthRegistration))

	// Authentication Verification
	mux.Handle(fmt.Sprintf("GET %s", routePaths.AuthVerification), p.baseMiddleware(p.handleGetAuthVerification))
	mux.Handle(fmt.Sprintf("GET %s", routePaths.AuthVerificationCode), p.baseMiddleware(p.handleGetAuthVerificationCode))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthVerificationEmail), p.baseMiddleware(p.handlePostVerificationEmail))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthVerificationCode), p.baseMiddleware(p.handlePostVerificationCode))

	// Authentication Login
	mux.Handle(fmt.Sprintf("GET %s", routePaths.AuthLogin), p.baseMiddleware(p.handleGetAuthLogin))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthLogin), p.baseMiddleware(p.handlePostAuthLogin))

	// Authentication Logout
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthLogout), p.baseMiddleware(p.handlePostAuthLogout))

	// Authentication Recovery
	mux.Handle(fmt.Sprintf("GET %s", routePaths.AuthRecovery), p.baseMiddleware(p.handleGetAuthRecovery))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthRecoveryEmail), p.baseMiddleware(p.handlePostAuthRecoveryEmail))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.AuthRecoveryCode), p.baseMiddleware(p.handlePostAuthRecoveryCode))

	// My Password
	mux.Handle(fmt.Sprintf("GET %s", routePaths.MyPassword), p.settingsMiddleware(p.handleGetMyPassword))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.MyPassword), p.settingsMiddleware(p.handlePostMyPassword))

	// My Profile
	mux.Handle(fmt.Sprintf("GET %s", routePaths.MyProfile), p.settingsMiddleware(p.handleGetMyProfile))
	mux.Handle(fmt.Sprintf("GET %s", routePaths.MyProfileEdit), p.settingsMiddleware(p.handleGetMyProfileEdit))
	mux.Handle(fmt.Sprintf("GET %s", routePaths.MyProfileForm), p.settingsMiddleware(p.handleGetMyProfileForm))
	mux.Handle(fmt.Sprintf("POST %s", routePaths.MyProfile), p.settingsMiddleware(p.handlePostMyProfile))

	// Top
	mux.Handle(fmt.Sprintf("GET %s", routePaths.Top), p.baseMiddleware(p.handleGetTop))

	// Item
	mux.Handle(fmt.Sprintf("GET %s", routePaths.Item), p.baseMiddleware(p.handleGetItemDetail))

	return mux
}

func (p *Provider) baseMiddleware(handler http.HandlerFunc) http.Handler {
	return p.loggingRquest(
		p.setSession(handler),
	)
}

func (p *Provider) settingsMiddleware(handler http.HandlerFunc) http.Handler {
	return p.loggingRquest(
		p.setSession(
			p.redirectIfExistsTraitsFieldsNotFilledIn(handler),
		),
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

func (p *Provider) redirectIfExistsTraitsFieldsNotFilledIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := getSession(ctx)

		isIgnoreEndpoint := slices.Contains([]string{
			fmt.Sprintf("POST %s", routePaths.MyProfile),
			fmt.Sprintf("GET %s", routePaths.MyProfileEdit),
			fmt.Sprintf("GET %s", routePaths.AuthLogin),
		}, fmt.Sprintf("%s %s", r.Method, r.URL.Path))

		if isAuthenticated(session) && existsTraitsFieldsNotFilledIn(session) && isIgnoreEndpoint {
			redirect(w, r, routePaths.MyProfileEdit)
		} else {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
