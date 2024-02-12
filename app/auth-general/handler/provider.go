package handler

import (
	"context"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"os"
	"slices"
)

type Provider struct {
	d Dependencies
}

type Dependencies struct {
	Kratos *kratos.Provider
}

type NewInput struct{}

func New(i NewInput, d Dependencies) (*Provider, error) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	p := Provider{
		d: d,
	}

	return &p, nil
}

func (p *Provider) baseMiddleware(hendler http.HandlerFunc) http.Handler {
	return p.setSession(
		p.redirectIfExistsTraitsFieldsNotFilledIn(hendler),
	)
}

func (p *Provider) settingsMiddleware(hendler http.HandlerFunc) http.Handler {
	return p.setSession(
		p.redirectIfExistsTraitsFieldsNotFilledIn(
			p.requireAuthenticated(hendler),
		),
	)
}

func (p *Provider) RegisterHandles(mux *http.ServeMux) *http.ServeMux {
	mux.Handle("GET /public/health", p.baseMiddleware(p.handleGetHealth))

	// // Registration
	mux.Handle("GET /registration", p.baseMiddleware(p.handleGetRegistration))
	mux.Handle("POST /registration", p.baseMiddleware(p.handlePostRegistration))

	// // Verification
	mux.Handle("GET /verification", p.baseMiddleware(p.handleGetVerification))
	mux.Handle("GET /verification/code", p.baseMiddleware(p.handleGetVerificationCode))
	mux.Handle("POST /verification/email", p.baseMiddleware(p.handlePostVerificationEmail))
	mux.Handle("POST /verification/code", p.baseMiddleware(p.handlePostVerificationCode))

	// Login
	mux.Handle("GET /login", p.baseMiddleware(p.handleGetLogin))
	mux.Handle("POST /login", p.baseMiddleware(p.handlePostLogin))

	// Logout
	mux.Handle("POST /logout", p.baseMiddleware(p.handlePostLogout))

	// Recovery
	mux.Handle("GET /recovery", p.baseMiddleware(p.handleGetRecovery))
	mux.Handle("POST /recovery/email", p.baseMiddleware(p.handlePostRecoveryEmail))
	mux.Handle("POST /recovery/code", p.baseMiddleware(p.handlePostRecoveryCode))

	// Settings
	mux.Handle("GET /settings/password", p.settingsMiddleware(p.handleGetPasswordSettings))
	mux.Handle("POST /settings/password", p.settingsMiddleware(p.handlePostSettingsPassword))
	mux.Handle("GET /settings/profile", p.settingsMiddleware(p.handleGetSettingsProfile))
	mux.Handle("GET /settings/profile/edit", p.settingsMiddleware(p.handleGetSettingsProfileEdit))
	mux.Handle("GET /settings/profile/_form", p.settingsMiddleware(p.handleGetSettingsProfileForm))
	mux.Handle("POST /settings/profile", p.settingsMiddleware(p.handlePostSettingsProfile))

	mux.Handle("GET /", p.baseMiddleware(p.handleGetHome))
	mux.Handle("GET /item/{id}", p.baseMiddleware(p.handleGetItemDetail))

	return mux
}

func (p *Provider) setSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("setSession")
		ctx := r.Context()
		output, err := p.d.Kratos.ToSession(kratos.ToSessionInput{
			Cookie: r.Header.Get("Cookie"),
		})
		if err != nil {
			slog.Info("setSession error")
			slog.Info(err.Error())
			ctx = context.WithValue(ctx, "session", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, "session", output.Session)
		session := getSession(ctx)
		slog.Info(fmt.Sprintf("%v", session))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (p *Provider) redirectIfExistsTraitsFieldsNotFilledIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := getSession(ctx)
		targetPaths := []string{"/settings/profile/edit", "/login"}
		slog.Info("redirectIfExistsTraitsFieldsNotFilledIn")
		slog.Info(fmt.Sprintf("%v", isAuthenticated(session)))
		slog.Info(fmt.Sprintf("%v", existsTraitsFieldsNotFilledIn(session)))
		if isAuthenticated(session) &&
			existsTraitsFieldsNotFilledIn(session) &&
			!slices.Contains(targetPaths, r.URL.Path) {
			redirect(w, r, fmt.Sprintf("%s/settings/profile/edit", generalEndpoint))
		} else {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (p *Provider) requireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := getSession(ctx)
		if isAuthenticated(session) {
			slog.Info("Authenticated")
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			slog.Info("Not Authenticated")
			redirect(w, r, fmt.Sprintf("%s/error/unauthorized", generalEndpoint))
		}
	})
}
