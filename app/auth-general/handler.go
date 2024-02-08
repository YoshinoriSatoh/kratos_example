package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	kratosclientgo "github.com/ory/kratos-client-go"
)

// Registration flowの作成とフォームレンダリング
func handleGetRegistration(c *gin.Context) {
	var (
		err              error
		uiErrorMessage   string
		response         *http.Response
		registrationFlow *kratosclientgo.RegistrationFlow
		csrfToken        string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	flowID := c.Query("flow")

	// flowID がない場合は新規にRegistration Flow を作成してリダイレクト
	if flowID == "" {
		registrationFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserRegistrationFlow(c).
			Execute()
		if err != nil {
			slog.Error("Create Registration Flow Error", "RegistrationFlow", registrationFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "registration/index.html", gin.H{
				"Title":        "Registration",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateBrowserRegistrationFlow Succeed", "RegistrationFlow", registrationFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/registration?flow=%s", generalEndpoint, registrationFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	registrationFlow, response, err = kratosPublicClient.FrontendApi.
		GetRegistrationFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Get Registration Flow Error", "RegistrationFlow", registrationFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "registration/index.html", gin.H{
			"Title":        "Registration",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetRegisrationFlow Succeed", "RegistrationFlow", registrationFlow, "Response", response)

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "registration/index.html", gin.H{
		"Title":              "Registration",
		"RegistrationFlowID": registrationFlow.Id,
		"CsrfToken":          csrfToken,
		"ErrorMessage":       uiErrorMessage,
	})
}

// Registration flowの送信(完了)と検証メールが送信された旨のメッセージレンダリング
func handlePostRegistrationForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	email := c.PostForm("email")
	password := c.PostForm("password")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Email", email, "Password", password, "CsrfToken", csrfToken, "cookie", cookie)

	// Registration Flow の送信(完了)
	updateRegistrationFlowBody := kratosclientgo.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &kratosclientgo.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
			Traits: map[string]interface{}{
				"email": email,
			},
			CsrfToken: &csrfToken,
		},
	}
	successfulRegistration, response, err := kratosPublicClient.FrontendApi.
		UpdateRegistrationFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateRegistrationFlowBody(updateRegistrationFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Registration Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "registration/_form.html", gin.H{
			"Title":        "Verification",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateRegisration Succeed", "SuccessfulRegistration", successfulRegistration, "Response", response)

	verificationFlowID, err := getContinueWithVerificationUiFlowIdFromFlowHttpResponse(response)
	if err != nil {
		slog.Error("Update Registration Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "registration/_form.html", gin.H{
			"Title":        "Verification",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// Registration flow成功時はVerification flowへリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, verificationFlowID))
	c.Status(200)
}

// Verification flowのフォームレンダリング
func handleGetVerification(c *gin.Context) {
	var (
		err              error
		uiErrorMessage   string
		response         *http.Response
		verificationFlow *kratosclientgo.VerificationFlow
		csrfToken        string
		flowAlreadyUsed  bool
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	// flowID がない場合は新規にVerification Flow を作成してリダイレクト
	flowID := c.Query("flow")
	if flowID == "" {
		verificationFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserVerificationFlow(c).
			Execute()
		if err != nil {
			slog.Error("Create Verification Flow Error", "VerificationFlow", verificationFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "verification/index.html", gin.H{
				"Title":        "Verification",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateVerificationFlow Succeed", "VerificationFlow", verificationFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, verificationFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	verificationFlow, response, err = kratosPublicClient.FrontendApi.
		GetVerificationFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Get Verification Flow Error", "VerificationFlow", verificationFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "verification/index.html", gin.H{
			"Title":        "Verification",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetVerificationFlow Succeed", "VerificationFlow", verificationFlow, "Response", response)

	// flow　が使用済みかチェック
	if verificationFlow.State == kratosclientgo.VERIFICATIONFLOWSTATE_PASSED_CHALLENGE {
		flowAlreadyUsed = true
		return
	}

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	c.HTML(http.StatusOK, "verification/index.html", gin.H{
		"Title":              "Verification",
		"VerificationFlowID": verificationFlow.Id,
		"CsrfToken":          csrfToken,
		"FlowAlreadyUsed":    flowAlreadyUsed,
		"ErrorMessage":       uiErrorMessage,
	})
}

// Verification flowの送信(完了)
func handlePostVerificationForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	code := c.PostForm("code")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Code", code, "CsrfToken", csrfToken)

	// Verification Flow の送信(完了)
	updateVerificationFlowBody := kratosclientgo.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &kratosclientgo.UpdateVerificationFlowWithCodeMethod{
			Method:    "code",
			Code:      &code,
			CsrfToken: &csrfToken,
		},
	}
	successfulVerification, response, err := kratosPublicClient.FrontendApi.
		UpdateVerificationFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateVerificationFlowBody(updateVerificationFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Verification Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "verification/_form.html", gin.H{
			"Title":        "Verification",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateVerification Succeed", "SuccessfulVerification", successfulVerification, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// Loign 画面へリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
	c.Status(200)
}

// Login flowのフォームレンダリング
func handleGetLogin(c *gin.Context) {
	var (
		err            error
		uiErrorMessage string
		response       *http.Response
		loginFlow      *kratosclientgo.LoginFlow
		csrfToken      string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	returnTo := c.Query("return_to")

	session, _, err := toSession(cookie)
	slog.Info(fmt.Sprintf("%v", session))
	if err == nil {
		for _, v := range session.AuthenticationMethods {
			slog.Info(*v.Method)
			if *v.Method == "code_recovery" {
				var logoutFlow *kratosclientgo.LogoutFlow
				logoutFlow, response, err := kratosPublicClient.FrontendApi.
					CreateBrowserLogoutFlow(c).
					Cookie(cookie).
					Execute()
				if err != nil {
					slog.Error("CreateLogoutFlow Error", "LogoutFlow", logoutFlow, "Response", response, "Error", err)
					c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
					c.Status(200)
					return
				}

				// Logout Flow の送信(完了)
				response, err = kratosPublicClient.FrontendApi.
					UpdateLogoutFlow(c).
					Token(logoutFlow.LogoutToken).
					Cookie(cookie).
					Execute()
				if err != nil {
					slog.Error("Update Logout Flow Error", "Response", response, "Error", err)
				} else {
					slog.Info("UpdateLoginFlow Succeed", "Response", response)
				}

				c.SetCookie("kratos_general_session", "", -1, "/", "localhost", false, true)
				c.Redirect(303, fmt.Sprintf("%s/login", generalEndpoint))
				return
			}
		}
		// c.Redirect(303, fmt.Sprintf("%s/", generalEndpoint))
		// return
	}

	flowID := c.Query("flow")

	// flowID がない場合は新規にLogin Flow を作成
	if flowID == "" {
		loginFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserLoginFlow(c).
			Refresh(true).
			Execute()
		if err != nil {
			slog.Error("CreateLoginFlow Error", "LoginFlow", loginFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "login/index.html", gin.H{
				"Title":        "Login",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateLoginFlow Succeed", "LoginFlow", loginFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/login?flow=%s", generalEndpoint, loginFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	loginFlow, response, err = kratosPublicClient.FrontendApi.
		GetLoginFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("GetLoginerr", "LoginFlow", loginFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "login/index.html", gin.H{
			"Title":        "Login",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetLoginFlow Succeed", "LoginFlow", loginFlow, "Response", response)

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	c.HTML(http.StatusOK, "login/index.html", gin.H{
		"Title":        "Login",
		"LoginFlowID":  loginFlow.Id,
		"ReturnTo":     returnTo,
		"CsrfToken":    csrfToken,
		"ErrorMessage": uiErrorMessage,
	})
}

// Login flowの送信(完了)
func handlePostLoginForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	returnTo := c.Query("return_to")
	csrfToken := c.PostForm("csrf_token")
	identifier := c.PostForm("identifier")
	password := c.PostForm("password")

	slog.Info("Params", "FlowID", flowID, "CsrfToken", csrfToken)

	// Login Flow の送信(完了)
	updateLoginFlowBody := kratosclientgo.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &kratosclientgo.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: identifier,
			Password:   password,
			CsrfToken:  &csrfToken,
		},
	}
	successfulLogin, response, err := kratosPublicClient.FrontendApi.
		UpdateLoginFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateLoginFlowBody(updateLoginFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Login Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "login/_form.html", gin.H{
			"Title":        "Login",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateLoginFlow Succeed", "SuccessfulLogin", successfulLogin, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// Login flow成功時はホーム画面へリダイレクト
	if returnTo != "" {
		c.Writer.Header().Set("HX-Redirect", returnTo)
	} else {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	}
	c.Status(200)
}

// Logout flowの送信(完了)
func handlePostLogout(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す

	var logoutFlow *kratosclientgo.LogoutFlow
	logoutFlow, response, err := kratosPublicClient.FrontendApi.
		CreateBrowserLogoutFlow(c).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("CreateLogoutFlow Error", "LogoutFlow", logoutFlow, "Response", response, "Error", err)
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
		c.Status(200)
		return
	}

	// Logout Flow の送信(完了)
	response, err = kratosPublicClient.FrontendApi.
		UpdateLogoutFlow(c).
		Token(logoutFlow.LogoutToken).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Update Logout Flow Error", "Response", response, "Error", err)
	} else {
		slog.Info("UpdateLoginFlow Succeed", "Response", response)
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// ログインセッションを削除
	c.SetCookie("kratos_general_session", "", -1, "/", "localhost", false, true)

	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	c.Status(200)
}

// Recovery flowの作成とフォームレンダリング
func handleGetRecovery(c *gin.Context) {
	var (
		err            error
		uiErrorMessage string
		response       *http.Response
		recoveryFlow   *kratosclientgo.RecoveryFlow
		csrfToken      string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	flowID := c.Query("flow")

	// flowID がない場合は新規にRecovery Flow を作成してリダイレクト
	if flowID == "" {
		recoveryFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserRecoveryFlow(c).
			Execute()
		if err != nil {
			slog.Error("Create Recovery Flow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "Recovery/index.html", gin.H{
				"Title":        "Recovery",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateBrowserRecoveryFlow Succeed", "RecoveryFlow", recoveryFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/recovery?flow=%s", generalEndpoint, recoveryFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	recoveryFlow, response, err = kratosPublicClient.FrontendApi.
		GetRecoveryFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Get Recovery Flow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "recovery/index.html", gin.H{
			"Title":        "Recovery",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetRecoveryFlow Succeed", "RecoveryFlow", recoveryFlow, "Response", response)

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "recovery/index.html", gin.H{
		"Title":          "Recovery",
		"RecoveryFlowID": recoveryFlow.Id,
		"CsrfToken":      csrfToken,
		"ErrorMessage":   uiErrorMessage,
	})
}

// Recovery flowの送信(完了)
func handlePostRecoveryCodeForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	email := c.PostForm("email")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Email", email, "CsrfToken", csrfToken, "cookie", cookie)

	// Recovery Flow の送信(完了)
	updateRecoveryFlowBody := kratosclientgo.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &kratosclientgo.UpdateRecoveryFlowWithCodeMethod{
			Method:    "code",
			Email:     &email,
			CsrfToken: &csrfToken,
		},
	}
	recoveryFlow, response, err := kratosPublicClient.FrontendApi.
		UpdateRecoveryFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateRecoveryFlowBody(updateRecoveryFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Recovery Flow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "recovery/_code_form.html", gin.H{
			"Title":        "Recovery",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateRecovery Succeed", "RecoveryFlow", recoveryFlow, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "recovery/_code_form.html", gin.H{
		"Title":          "Recovery",
		"RecoveryFlowID": flowID,
		"CsrfToken":      csrfToken,
	})
}

func completeRecoveryFlow(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	code := c.PostForm("code")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Code", code, "CsrfToken", csrfToken, "cookie", cookie)

	// Recovery Flow の送信(完了)
	updateRecoveryFlowBody := kratosclientgo.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &kratosclientgo.UpdateRecoveryFlowWithCodeMethod{
			Method:    "code",
			Code:      &code,
			CsrfToken: &csrfToken,
		},
	}
	recoveryFlow, response, err := kratosPublicClient.FrontendApi.
		UpdateRecoveryFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateRecoveryFlowBody(updateRecoveryFlowBody).
		Execute()
	if err != nil {
		if response.StatusCode == 422 {
			redirectBrowserTo, _ := getRedirectBrowserToFromFlowHttpResponse(response)
			setCookie(c, response)
			slog.Info(redirectBrowserTo)
			c.Writer.Header().Set("HX-Redirect", redirectBrowserTo)
			c.Status(200)
		} else {
			slog.Error("Update Recovery Flow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "recovery/_code_form.html", gin.H{
				"Title":        "Verification",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
		}
		return
	}
	c.Status(200)
}

// Settings
func handleGetPasswordSettings(c *gin.Context) {
	var (
		err            error
		uiErrorMessage string
		response       *http.Response
		settingsFlow   *kratosclientgo.SettingsFlow
		csrfToken      string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	session, _, err := toSession(cookie)

	flowID := c.Query("flow")

	// flowID がない場合は新規にSettings Flow を作成してリダイレクト
	if flowID == "" {
		settingsFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserSettingsFlow(c).
			Cookie(cookie).
			Execute()
		if err != nil {
			slog.Error("Create Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "settings/password.html", gin.H{
				"Title":        "Settings",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateBrowserSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/settings/password?flow=%s", generalEndpoint, settingsFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	settingsFlow, response, err = kratosPublicClient.FrontendApi.
		GetSettingsFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Get Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "settings/password.html", gin.H{
			"Title":        "Settings",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetRegisrationFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/password.html", gin.H{
		"Title":          "Settings",
		"SettingsFlowID": settingsFlow.Id,
		"CsrfToken":      csrfToken,
		"Session":        session,
		"ErrorMessage":   uiErrorMessage,
	})
}

func handlePostSettingsPasswordForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	password := c.PostForm("password")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Password", password, "CsrfToken", csrfToken, "cookie", cookie)

	// Settings Flow の送信(完了)
	updateSettingsFlowBody := kratosclientgo.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &kratosclientgo.UpdateSettingsFlowWithPasswordMethod{
			Method:    "password",
			Password:  password,
			CsrfToken: &csrfToken,
		},
	}
	successfulSettings, response, err := kratosPublicClient.FrontendApi.
		UpdateSettingsFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateSettingsFlowBody(updateSettingsFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Settings Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "settings/_password_form.html", gin.H{
			"Title":        "Settings",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateRegisration Succeed", "SuccessfulSettings", successfulSettings, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// Settings flow成功時はVerification flowへリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
	c.Status(200)
}

func handleGetProfileSettingsView(c *gin.Context) {
	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	session, _, _ := toSession(cookie)

	// flowID := c.Query("flow")

	// // flowID がない場合は新規にSettings Flow を作成してリダイレクト
	// if flowID == "" {
	// 	settingsFlow, response, err = kratosPublicClient.FrontendApi.
	// 		CreateBrowserSettingsFlow(c).
	// 		Cookie(cookie).
	// 		Execute()
	// 	if err != nil {
	// 		slog.Error("Create Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
	// 		c.HTML(http.StatusOK, "settings/profile_view.html", gin.H{
	// 			"Title":        "Settings",
	// 			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
	// 		})
	// 		return
	// 	}
	// 	slog.Info("CreateBrowserSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

	// 	setCookie(c, response)
	// 	c.Redirect(303, fmt.Sprintf("%s/settings/profile/view?flow=%s", generalEndpoint, settingsFlow.Id))
	// 	return
	// }

	// // flowID取得（CSRF Token の取得に必要）
	// settingsFlow, response, err = kratosPublicClient.FrontendApi.
	// 	GetSettingsFlow(context.Background()).
	// 	Id(flowID).
	// 	Cookie(cookie).
	// 	Execute()
	// if err != nil {
	// 	slog.Error("Get Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
	// 	c.HTML(http.StatusOK, "settings/profile_view.html", gin.H{
	// 		"Title":        "Settings",
	// 		"ErrorMessage": "Sorry, something went wrong. Please try again later.",
	// 	})
	// 	return
	// }
	// slog.Info("GetRegisrationFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

	// // flow の ui から csrf_token を取得
	// csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	return
	// }

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	// setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/profile_view.html", gin.H{
		"Title": "Profile",
		// "SettingsFlowID": settingsFlow.Id,
		// "CsrfToken":      csrfToken,
		"Session": session,
		// "ErrorMessage": uiErrorMessage,
	})
}

func handleGetProfileSettingsEdit(c *gin.Context) {
	var (
		err          error
		response     *http.Response
		settingsFlow *kratosclientgo.SettingsFlow
		csrfToken    string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	session, _, _ := toSession(cookie)

	flowID := c.Query("flow")

	// flowID がない場合は新規にSettings Flow を作成してリダイレクト
	if flowID == "" {
		settingsFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserSettingsFlow(c).
			Cookie(cookie).
			Execute()
		if err != nil {
			slog.Error("Create Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "settings/profile_view.html", gin.H{
				"Title":        "Settings",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateBrowserSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

		setCookie(c, response)
		c.Redirect(303, fmt.Sprintf("%s/settings/profile/edit?flow=%s", generalEndpoint, settingsFlow.Id))
		return
	}

	// flowID取得（CSRF Token の取得に必要）
	settingsFlow, response, err = kratosPublicClient.FrontendApi.
		GetSettingsFlow(context.Background()).
		Id(flowID).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("Get Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
		c.HTML(http.StatusOK, "settings/profile_form.html", gin.H{
			"Title":        "Settings",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("GetRegisrationFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/profile_edit.html", gin.H{
		"Title":          "Profile",
		"SettingsFlowID": settingsFlow.Id,
		"CsrfToken":      csrfToken,
		"Session":        session,
	})
}

func handleGetProfileSettingsForm(c *gin.Context) {
	var (
		err          error
		response     *http.Response
		settingsFlow *kratosclientgo.SettingsFlow
		csrfToken    string
	)

	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	session, _, _ := toSession(cookie)
	flowID := c.Query("flow")

	// flowID がない場合は新規にSettings Flow を作成してリダイレクト
	if flowID == "" {
		settingsFlow, response, err = kratosPublicClient.FrontendApi.
			CreateBrowserSettingsFlow(c).
			Cookie(cookie).
			Execute()
		if err != nil {
			slog.Error("Create Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "settings/profile_view.html", gin.H{
				"Title":        "Settings",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("CreateBrowserSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

		// setCookie(c, response)
		// // c.Redirect(303, fmt.Sprintf("%s/settings/profile/view?flow=%s", generalEndpoint, settingsFlow.Id))
		// c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/settings/profile/view_form?flow=%s", generalEndpoint, settingsFlow.Id))
		// c.Status(200)
		// return
	} else {

		// flowID取得（CSRF Token の取得に必要）
		settingsFlow, response, err = kratosPublicClient.FrontendApi.
			GetSettingsFlow(context.Background()).
			Id(flowID).
			Cookie(cookie).
			Execute()
		if err != nil {
			slog.Error("Get Settings Flow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			c.HTML(http.StatusOK, "settings/_profile_form.html", gin.H{
				"Title":        "Settings",
				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
			})
			return
		}
		slog.Info("GetRegisrationFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	authenticateAt := session.AuthenticatedAt.In(jst)
	slog.Info(fmt.Sprintf("%v", authenticateAt))
	slog.Info(fmt.Sprintf("%v", time.Now().Add(time.Minute*10)))
	if authenticateAt.After(time.Now().Add(time.Minute * 10)) {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, "/settings/profile_edit?flow="+flowID))
		c.Status(200)
	}

	// flow の ui から csrf_token を取得
	csrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/_profile_form.html", gin.H{
		"Title":          "Profile",
		"SettingsFlowID": settingsFlow.Id,
		"CsrfToken":      csrfToken,
		"Session":        session,
	})
}

func handlePostSettingsProfileForm(c *gin.Context) {
	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	flowID := c.Query("flow")
	nickname := c.PostForm("nickname")
	birthdate := c.PostForm("birthdate")
	csrfToken := c.PostForm("csrf_token")

	slog.Info("Params", "FlowID", flowID, "Nickname", nickname, "Birthdate", birthdate, "CsrfToken", csrfToken, "cookie", cookie)

	var email string

	session, _, err := toSession(cookie)
	slog.Info(fmt.Sprintf("%v", session))
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	authenticateAt := session.AuthenticatedAt.In(jst)
	slog.Info(fmt.Sprintf("%v", authenticateAt))
	slog.Info(fmt.Sprintf("%v", time.Now().Add(time.Minute*10)))
	if authenticateAt.Before(time.Now().Add(time.Minute * 10)) {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, "/settings/profile_edit?flow="+flowID))
		c.Status(200)
	}

	if err == nil {
		email = session.Identity.Traits.(map[string]interface{})["email"].(string)
	}

	if nickname == "" && session.Identity.Traits.(map[string]interface{})["nickname"] != nil {
		nickname = session.Identity.Traits.(map[string]interface{})["nickname"].(string)
	}

	if birthdate == "" && session.Identity.Traits.(map[string]interface{})["birthdate"] != nil {
		birthdate = session.Identity.Traits.(map[string]interface{})["birthdate"].(string)
	}

	// Settings Flow の送信(完了)
	updateSettingsFlowBody := kratosclientgo.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithProfileMethod: &kratosclientgo.UpdateSettingsFlowWithProfileMethod{
			Method: "profile",
			Traits: map[string]interface{}{
				"email":     email,
				"nickname":  nickname,
				"birthdate": birthdate,
			},
			CsrfToken: &csrfToken,
		},
	}
	successfulSettings, response, err := kratosPublicClient.FrontendApi.
		UpdateSettingsFlow(c).
		Flow(flowID).
		Cookie(cookie).
		UpdateSettingsFlowBody(updateSettingsFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Settings Flow Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "settings/_profile_form.html", gin.H{
			"Title":        "Settings",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("UpdateRegisration Succeed", "SuccessfulSettings", successfulSettings, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	// Settings flow成功時はVerification flowへリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	c.Status(200)
}

// Home画面（ログイン必須）レンダリング
func handleGetHome(c *gin.Context) {
	// browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す
	cookie := c.Request.Header.Get("Cookie")

	session, response, err := toSession(cookie)
	if err != nil && response.StatusCode != http.StatusUnauthorized {
		slog.Error("ToSession Error", "Response", response, "Error", err)
		c.HTML(http.StatusOK, "home/_form.html", gin.H{
			"Title":        "Login",
			"ErrorMessage": "Sorry, something went wrong. Please try again later.",
		})
		return
	}
	slog.Info("ToSession Result", "Session", session, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	setCookie(c, response)

	c.HTML(http.StatusOK, "home/index.html", gin.H{
		"Title":   "Home",
		"Session": session,
	})
}

// // My menuのレンダリング
// func handleGetMyMenu(c *gin.Context) {
// 	var err error

// 	cookie := c.Request.Header.Get("Cookie") // browser flowでは、ブラウザのcookieをそのままkratosへ受け渡す

// 	session, response, err := toSession(cookie)
// 	if err != nil {
// 		// 未認証の場合はログイン画面へリダイレクト
// 		if response.StatusCode == http.StatusUnauthorized {
// 			c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
// 			c.Status(200)
// 		} else {
// 			slog.Error("ToSession Error", "Response", response, "Error", err)
// 			c.HTML(http.StatusOK, "my/_menu.html", gin.H{
// 				"Title":        "Login",
// 				"ErrorMessage": "Sorry, something went wrong. Please try again later.",
// 			})
// 		}
// 		return
// 	}

// 	var logoutFlow *kratosclientgo.LogoutFlow
// 	logoutFlow, response, err = kratosPublicClient.FrontendApi.
// 		CreateBrowserLogoutFlow(c).
// 		Cookie(cookie).
// 		Execute()
// 	if err != nil {
// 		slog.Error("CreateLogoutFlow Error", "LogoutFlow", logoutFlow, "Response", response, "Error", err)
// 		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
// 		c.Status(200)
// 		return
// 	}

// 	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
// 	setCookie(c, response)

// 	c.HTML(http.StatusOK, "my/_menu.html", gin.H{
// 		"Session":     session,
// 		"LogoutToken": logoutFlow.LogoutToken,
// 	})
// }

func toSession(cookie string) (*kratosclientgo.Session, *http.Response, error) {
	session, response, err := kratosPublicClient.FrontendApi.
		ToSession(context.Background()).
		Cookie(cookie).
		Execute()
	if err != nil {
		slog.Error("ToSession Error", "Response", response, "Error", err)
	}

	return session, response, err
}
