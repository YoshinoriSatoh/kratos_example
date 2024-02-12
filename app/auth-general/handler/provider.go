package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/gin-gonic/gin"
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

func (p *Provider) Register(e *gin.Engine) {
	e.Use(p.getKratosSession())
	e.Use(p.redirectIfExistsTraitsFieldsNotFilledIn())

	e.GET("/public/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Registration
	e.GET("/registration", p.handleGetRegistration)
	e.POST("/registration", p.handlePostRegistration)

	// Verification
	e.GET("/verification", p.handleGetVerification)
	e.GET("/verification/code", p.handleGetVerificationCode)
	e.POST("/verification/email", p.handlePostVerificationEmail)
	e.POST("/verification/code", p.handlePostVerificationCode)

	// Login
	e.GET("/login", p.handleGetLogin)
	e.POST("/login", p.handlePostLogin)

	// Logout
	e.POST("/logout", p.handlePostLogout)

	// Recovery
	e.GET("/recovery", p.handleGetRecovery)
	e.POST("/recovery/email", p.handlePostRecoveryEmail)
	e.POST("/recovery/code", p.handlePostRecoveryCode)

	// Settings
	settingsRoute := e.Group("/settings", p.requireAuthenticated())
	settingsRoute.GET("/password", p.handleGetPasswordSettings)
	settingsRoute.POST("/password", p.handlePostSettingsPassword)
	settingsRoute.GET("/profile", p.handleGetSettingsProfile)
	settingsRoute.GET("/profile/edit", p.handleGetSettingsProfileEdit)
	settingsRoute.GET("/profile/_form", p.handleGetSettingsProfileForm)
	settingsRoute.POST("/profile", p.handlePostSettingsProfile)

	e.GET("/", p.handleGetHome)
	e.GET("/item/:id", p.handleGetItemDetail)
}

func (p *Provider) getKratosSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		output, err := p.d.Kratos.ToSession(kratos.ToSessionInput{
			Cookie: c.Request.Header.Get("Cookie"),
		})
		if err != nil {
			c.Set("session", nil)
			return
		}
		c.Set("session", output.Session)

		c.Next()
	}
}

func (p *Provider) redirectIfExistsTraitsFieldsNotFilledIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := getSession(c)
		targetPaths := []string{"/settings/profile/edit", "/login"}
		if isAuthenticated(c) &&
			existsTraitsFieldsNotFilledIn(session) &&
			!slices.Contains(targetPaths, c.Request.URL.Path) {
			c.Redirect(303, fmt.Sprintf("%s/settings/profile/edit", generalEndpoint))
		} else {
			c.Next()
		}
	}
}

func (p *Provider) requireAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAuthenticated(c) {
			slog.Info("Authenticated")
			c.Next()
		} else {
			slog.Info("Not Authenticated")
			// c.AbortWithStatus(http.StatusUnauthorized)
			c.Redirect(303, fmt.Sprintf("%s/error/unauthorized", generalEndpoint))
		}
	}
}
