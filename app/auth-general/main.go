package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	kratosclientgo "github.com/ory/kratos-client-go"
)

var generalEndpoint string = "http://localhost:3000"
var kratosPublicEndpoint string = "http://kratos-general:4433"
var kratosPublicClient *kratosclientgo.APIClient

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	kratosPublicConfigration := kratosclientgo.NewConfiguration()
	kratosPublicConfigration.Servers = []kratosclientgo.ServerConfiguration{{URL: kratosPublicEndpoint}}
	kratosPublicClient = kratosclientgo.NewAPIClient(kratosPublicConfigration)
}

func setCookie(c *gin.Context, response *http.Response) {
	for _, cookie := range response.Header["Set-Cookie"] {
		c.Writer.Header().Add("Set-Cookie", cookie)
	}
}

func main() {
	e := gin.New()
	e.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/public/health"},
	}))
	e.Use(gin.Recovery())

	e.LoadHTMLGlob("templates/**/*.html")
	e.Static("/static", "./static")

	e.GET("/public/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Registration
	e.GET("/registration", handleGetRegistration)
	e.POST("/registration/form", handlePostRegistrationForm)

	// Verification
	e.GET("/verification", handleGetVerification)
	e.POST("/verification/form", handlePostVerificationForm)

	// Login
	e.GET("/login", handleGetLogin)
	e.POST("/login/form", handlePostLoginForm)

	// Logout
	e.POST("/logout", handlePostLogout)

	// Recovery
	e.GET("/recovery", handleGetRecovery)
	e.POST("/recovery/code_form", handlePostRecoveryCodeForm)
	e.POST("/recovery/complete", completeRecoveryFlow)

	// Settings
	e.GET("/settings/password", handleGetPasswordSettings)
	e.POST("/settings/password_form", handlePostSettingsPasswordForm)
	e.GET("/settings/profile/view", handleGetProfileSettingsView)
	e.GET("/settings/profile/edit", handleGetProfileSettingsEdit)
	e.GET("/settings/profile_form", handleGetProfileSettingsForm)
	e.POST("/settings/profile_form", handlePostSettingsProfileForm)

	e.GET("/", handleGetHome)

	e.Run(":3000")
}

// flow の ui から csrf_token を取得
// SDKを使用しているので、本来は上記レスポンスの第一引数である registrationFlow *kratosclientgo.RegistrationFlow から取得するところだが、
// goのv1.0.0のSDKには不具合があるらしく、仕方ないのでhttp.Responseから取得している
// https://github.com/ory/sdk/issues/292
func getCsrfTokenFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Can not unmarshal JSON")
		return "", err
	}

	var csrfToken string
	for _, node := range result.(map[string]interface{})["ui"].(map[string]interface{})["nodes"].([]interface{}) {
		attrName := node.(map[string]interface{})["attributes"].(map[string]interface{})["name"]
		if attrName != nil && attrName.(string) == "csrf_token" {
			csrfToken = node.(map[string]interface{})["attributes"].(map[string]interface{})["value"].(string)
			break
		}
	}
	slog.Info(csrfToken)
	return csrfToken, nil
}

func getContinueWithVerificationUiFlowIdFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Can not unmarshal JSON")
		return "", err
	}

	var verificationFlowID string
	for _, continueWith := range result.(map[string]interface{})["continue_with"].([]interface{}) {
		if continueWith.(map[string]interface{})["action"].(string) == "show_verification_ui" {
			verificationFlowID = continueWith.(map[string]interface{})["flow"].(map[string]interface{})["id"].(string)
			break
		}
	}
	slog.Info(verificationFlowID)
	return verificationFlowID, nil
}

func getRedirectBrowserToFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Can not unmarshal JSON")
		return "", err
	}
	return result.(map[string]interface{})["redirect_browser_to"].(string), nil
}
