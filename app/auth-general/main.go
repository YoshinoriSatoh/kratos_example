package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	kratosclientgo "github.com/ory/kratos-client-go"
)

var (
	generalEndpoint              string = "http://localhost:3000"
	kratosPublicEndpoint         string = "http://kratos-general:4433"
	kratosPublicClient           *kratosclientgo.APIClient
	locationJst                  *time.Location
	privilegedAccessLimitMinutes time.Duration = 10
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	kratosPublicConfigration := kratosclientgo.NewConfiguration()
	kratosPublicConfigration.Servers = []kratosclientgo.ServerConfiguration{{URL: kratosPublicEndpoint}}
	kratosPublicClient = kratosclientgo.NewAPIClient(kratosPublicConfigration)

	var err error
	locationJst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

func main() {
	e := gin.New()
	e.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/public/health"},
	}))
	e.Use(gin.Recovery())

	templateList := template.Must(template.New("").ParseGlob("templates/**/*.html"))
	templateList = template.Must(templateList.ParseGlob("templates/**/**/*.html"))

	e.SetHTMLTemplate(templateList)
	e.Static("/static", "./static")

	e.Use(getKratosSession())
	e.Use(redirectIfExistsTraitsFieldsNotFilledIn())

	e.GET("/public/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Registration
	e.GET("/registration", handleGetRegistration)
	e.POST("/registration", handlePostRegistration)

	// Verification
	e.GET("/verification", handleGetVerification)
	e.GET("/verification/code", handleGetVerificationCode)
	e.POST("/verification/email", handlePostVerificationEmail)
	e.POST("/verification/code", handlePostVerificationCode)

	// Login
	e.GET("/login", handleGetLogin)
	e.POST("/login", handlePostLogin)

	// Logout
	e.POST("/logout", handlePostLogout)

	// Recovery
	e.GET("/recovery", handleGetRecovery)
	e.POST("/recovery/email", handlePostRecoveryEmail)
	e.POST("/recovery/code", handlePostRecoveryCode)

	// Settings
	settingsRoute := e.Group("/settings", requireAuthenticated())
	settingsRoute.GET("/password", handleGetPasswordSettings)
	settingsRoute.POST("/password", handlePostSettingsPassword)
	settingsRoute.GET("/profile", handleGetSettingsProfile)
	settingsRoute.GET("/profile/edit", handleGetSettingsProfileEdit)
	settingsRoute.GET("/profile/_form", handleGetSettingsProfileForm)
	settingsRoute.POST("/profile", handlePostSettingsProfile)

	e.GET("/", handleGetHome)
	e.GET("/item/:id", handleGetItemDetail)

	e.Run(":3000")
}
