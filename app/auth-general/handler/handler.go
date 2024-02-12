package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Handler GET /registration
func (p *Provider) handleGetRegistration(c *gin.Context) {
	// Registration Flow の作成 or 取得
	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRegistrationFlow(kratos.CreateOrGetRegistrationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "registration/index.html", viewParameters(c, gin.H{
			"Title":         "Registration",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/registration?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "registration/index.html", viewParameters(c, gin.H{
		"Title":              "Registration",
		"RegistrationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
	}))
}

// Handler POST /registration
func (p *Provider) handlePostRegistration(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     c.PostForm("email"),
		Password:  c.PostForm("password"),
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "registration/_form.html", viewParameters(c, gin.H{
			"Title":              "Registration",
			"RegistrationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Registration flow成功時はVerification flowへリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/verification/code?flow=%s", generalEndpoint, output.VerificationFlowID))
	c.Status(200)
}

// Handler GET /verification handler
func (p *Provider) handleGetVerification(c *gin.Context) {
	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/index.html", viewParameters(c, gin.H{
			"Title":         "Verification",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	c.HTML(http.StatusOK, "verification/index.html", viewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler GET /verification/code handler
func (p *Provider) handleGetVerificationCode(c *gin.Context) {
	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/code.html", viewParameters(c, gin.H{
			"Title":         "Verification",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	c.HTML(http.StatusOK, "verification/code.html", viewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler POST /verification/email
func (p *Provider) handlePostVerificationEmail(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	email := c.PostForm("email")

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     email,
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/_code_form.html", viewParameters(c, gin.H{
			"Title":              "Verification",
			"VerificationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	c.HTML(http.StatusOK, "verification/_code_form.html", viewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": flowID,
		"CsrfToken":          csrfToken,
		"ErrorMessages":      output.ErrorMessages,
	}))
}

// Handler POST /verification/code
func (p *Provider) handlePostVerificationCode(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	code := c.PostForm("code")

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Code:      code,
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/_code_form.html", viewParameters(c, gin.H{
			"Title":              "Verification",
			"VerificationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Loign 画面へリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
	c.Status(200)
}

// Handler GET /login
func (p *Provider) handleGetLogin(c *gin.Context) {
	var refresh bool

	// 認証済みの場合、認証時刻の更新を実施
	// プロフィール設定時に認証時刻が一定期間内である必要があり、過ぎている場合はログイン画面へリダイレクトし、ログインを促している
	if isAuthenticated(c) {
		refresh = true
	}

	// Login Flow の作成 or 取得
	output, err := p.d.Kratos.CreateOrGetLoginFlow(kratos.CreateOrGetLoginFlowInput{
		Cookie:  c.Request.Header.Get("Cookie"),
		FlowID:  c.Query("flow"),
		Refresh: refresh,
	})
	if err != nil {
		c.HTML(http.StatusOK, "login/index.html", viewParameters(c, gin.H{
			"Title":         "Login",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Login flowを新規作成した場合は、FlowIDを含めてリダイレクト
	// return_to
	//   指定時: ログイン後にreturn_toで指定されたURLへリダイレクト
	//   未指定時: ログイン後にホーム画面へリダイレクト
	returnTo := url.QueryEscape(c.Query("return_to"))
	slog.Info(returnTo)
	slog.Info(url.QueryEscape("/settings/profile"))
	if output.IsNewFlow {
		var redirectTo string
		if returnTo == "" {
			redirectTo = fmt.Sprintf("%s/login?flow=%s", generalEndpoint, output.FlowID)
		} else {
			redirectTo = fmt.Sprintf("%s/login?flow=%s&return_to=%s", generalEndpoint, output.FlowID, returnTo)
		}
		redirect(c, redirectTo)
		return
	}

	var information string
	if existsAfterLoginHook(c, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィール更新のために、再度ログインをお願いします。"
	}

	c.HTML(http.StatusOK, "login/index.html", viewParameters(c, gin.H{
		"Title":       "Login",
		"LoginFlowID": output.FlowID,
		"ReturnTo":    returnTo,
		"Information": information,
		"CsrfToken":   output.CsrfToken,
	}))
}

// Handler POST /login
func (p *Provider) handlePostLogin(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Login Flow 更新
	output, err := p.d.Kratos.UpdateLoginFlow(kratos.UpdateLoginFlowInput{
		Cookie:     c.Request.Header.Get("Cookie"),
		FlowID:     flowID,
		CsrfToken:  csrfToken,
		Identifier: c.PostForm("identifier"),
		Password:   c.PostForm("password"),
	})
	if err != nil {
		c.Status(400)
		c.HTML(http.StatusOK, "login/_form.html", viewParameters(c, gin.H{
			"Title":         "Login",
			"LoginFlowID":   flowID,
			"CsrfToken":     csrfToken,
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// ログインフック実行
	hook, err := loadAfterLoginHook(c, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprintf("%v", hook))
	slog.Info(fmt.Sprintf("%v", hook))
	if hook.Operation == AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE {
		hookParams, _ := hook.Params.(map[string]interface{})
		err := p.updateProfile(c, settingsProfileEditViewParams{
			FlowID:    hookParams["flow_id"].(string),
			Email:     hookParams["email"].(string),
			Nickname:  hookParams["nickname"].(string),
			Birthdate: hookParams["birthdate"].(string),
		})
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}

	// Login flow成功時:
	//   Traitsに設定させたい項目がまだ未設定の場合、Settings(profile)へリダイレクト
	//   はホーム画面へリダイレクト
	returnTo := c.Query("return_to")
	slog.Info(returnTo)
	var redirectTo string
	if returnTo != "" {
		redirectTo = returnTo
	} else {
		redirectTo = fmt.Sprintf("%s/", generalEndpoint)
	}
	redirect(c, redirectTo)
}

// Handler POST /logout
func (p *Provider) handlePostLogout(c *gin.Context) {
	// Logout
	_, err := p.d.Kratos.Logout(kratos.LogoutFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
	})
	c.SetCookie("kratos_general_session", "", -1, "/", "localhost", false, true)
	if err != nil {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
		c.Status(200)
		return
	}

	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	c.Status(200)
}

// Handler GET /recovery
func (p *Provider) handleGetRecovery(c *gin.Context) {
	// Recovery Flow の作成 or 取得
	// Recovery flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRecoveryFlow(kratos.CreateOrGetRecoveryFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "recovery/index.html", viewParameters(c, gin.H{
			"Title":         "Recovery",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/recovery?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "recovery/index.html", viewParameters(c, gin.H{
		"Title":          "Recovery",
		"RecoveryFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
	}))
}

// Handler POST /recovery/email
func (p *Provider) handlePostRecoveryEmail(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Email:     c.PostForm("email"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "recovery/_code_form.html", viewParameters(c, gin.H{
			"Title":          "Recovery",
			"RecoveryFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "recovery/_code_form.html", viewParameters(c, gin.H{
		"Title":                    "Recovery",
		"RecoveryFlowID":           flowID,
		"CsrfToken":                csrfToken,
		"ShowRecoveryAnnouncement": true,
	}))
}

// Handler POST /recovery/code
func (p *Provider) handlePostRecoveryCode(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Code:      c.PostForm("code"),
	})
	if err != nil && output.RedirectBrowserTo == "" {
		c.HTML(http.StatusOK, "recovery/_code_form.html", viewParameters(c, gin.H{
			"Title":          "Recovery",
			"RecoveryFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s&from=recovery", output.RedirectBrowserTo))
	c.Status(200)
}

// Handler GET /settings/password
func (p *Provider) handleGetPasswordSettings(c *gin.Context) {
	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})

	if err != nil {
		c.HTML(http.StatusOK, "settings/password/index.html", viewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/password/index.html", viewParameters(c, gin.H{
		"Title":                "Settings",
		"SettingsFlowID":       output.FlowID,
		"CsrfToken":            output.CsrfToken,
		"RedirectFromRecovery": c.Query("from") == "recovery",
	}))
}

// Handler POST /settings/password
func (p *Provider) handlePostSettingsPassword(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Setting Flow 更新
	output, err := p.d.Kratos.UpdateSettingsFlowPassword(kratos.UpdateSettingsFlowPasswordInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Password:  c.PostForm("password"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/password/_form.html", viewParameters(c, gin.H{
			"Title":          "Settings",
			"SettingsFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Settings flow成功時はVerification flowへリダイレクト
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login", generalEndpoint))
	c.Status(200)
}

// Handler GET /settings/profile
func (p *Provider) handleGetSettingsProfile(c *gin.Context) {
	session := getSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/index.html", viewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// flowの情報に従ってレンダリング
	email := session.GetValueFromTraits("email")
	nickname := session.GetValueFromTraits("nickname")
	birthdate := session.GetValueFromTraits("birthdate")
	var information string
	if existsAfterLoginHook(c, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィールを更新しました。"
		deleteAfterLoginHook(c, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	}
	c.HTML(http.StatusOK, "settings/profile/index.html", viewParameters(c, gin.H{
		"Title":          "Profile",
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          email,
		"Nickname":       nickname,
		"Birthdate":      birthdate,
		"Information":    information,
	}))
}

// Handler GET /settings/profile/edit
func (p *Provider) handleGetSettingsProfileEdit(c *gin.Context) {
	session := getSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/edit.html", viewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// セッションから現在の値を取得
	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{}, session)

	var information string
	if existsTraitsFieldsNotFilledIn(session) {
		information = "プロフィールの入力をお願いします"
	}

	c.HTML(http.StatusOK, "settings/profile/edit.html", viewParameters(c, gin.H{
		"Title":          "Profile",
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          params.Email,
		"Nickname":       params.Nickname,
		"Birthdate":      params.Birthdate,
		"Infomation":     information,
	}))
}

// Handler GET /settings/profile/_form
func (p *Provider) handleGetSettingsProfileForm(c *gin.Context) {
	session := getSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/_form.html", viewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// // セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	// if NeedLoginWhenPrivilegedAccess(c) {
	// 	returnTo := url.QueryEscape("/settings/profile/edit")
	// 	redirectTo := fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, returnTo)
	// 	redirect(c, redirectTo)
	// 	return
	// }

	// セッションから現在の値を取得
	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{}, session)

	c.HTML(http.StatusOK, "settings/profile/_form.html", viewParameters(c, gin.H{
		"Title":          "Profile",
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          params.Email,
		"Nickname":       params.Nickname,
		"Birthdate":      params.Birthdate,
	}))
}

// Handler POST /settings/profile
type handlePostSettingsProfileRequestPostForm struct {
	Email     string
	Nickname  string
	Birthdate string
}

func (f handlePostSettingsProfileRequestPostForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Nickname, validation.Required, validation.Length(5, 20)),
		validation.Field(&f.Birthdate, validation.Required, validation.Date("2006-01-02")),
	)
}

// Handler POST /settings/profile
func (p *Provider) handlePostSettingsProfile(c *gin.Context) {
	session := getSession(c)
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	requestPostForm := handlePostSettingsProfileRequestPostForm{
		Email:     c.PostForm("email"),
		Nickname:  c.PostForm("nickname"),
		Birthdate: c.PostForm("birthdate"),
	}
	slog.Info("requestPostForm", "nickname", requestPostForm.Nickname, "birthdate", requestPostForm.Birthdate)

	err := requestPostForm.Validate()
	if err != nil {
		validationErrors := make(map[string]string, 3)
		errs := err.(validation.Errors)
		for k, err := range errs {
			validationErrors[k] = err.Error()
			slog.Info(fmt.Sprintf("%v", err.(validation.Error)))
			verr := err.(validation.Error)
			slog.Info(verr.Error())
			slog.Info(verr.Code())
			slog.Info(verr.Message())
		}
		c.HTML(http.StatusOK, "settings/profile/_form.html", viewParameters(c, gin.H{
			"SettingsFlowID":  flowID,
			"CsrfToken":       csrfToken,
			"ErrorMessages":   []string{"Error"},
			"Email":           requestPostForm.Email,
			"Nickname":        requestPostForm.Nickname,
			"Birthdate":       requestPostForm.Birthdate,
			"ValidationError": validationErrors,
		}))
		return
	}

	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{
		FlowID:    flowID,
		Email:     requestPostForm.Email,
		Nickname:  requestPostForm.Nickname,
		Birthdate: requestPostForm.Birthdate,
	}, session)
	slog.Info("params", "email", params.Email, "nickname", params.Nickname, "birthdate", params.Birthdate)

	deleteAfterLoginHook(c, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if session.NeedLoginWhenPrivilegedAccess(c) {
		slog.Info(fmt.Sprintf("%v", params))
		// err := saveSettingsProfileEditParamsToCookie(c, params)
		err := saveAfterLoginHook(c, afterLoginHook{
			Operation: AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE,
			Params:    params,
		}, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
		if err != nil {
			c.HTML(http.StatusOK, "settings/profile/_form.html", viewParameters(c, gin.H{
				"Title":          "Settings",
				"SettingsFlowID": flowID,
				"CsrfToken":      csrfToken,
				"ErrorMessages":  []string{"Error"},
				"Email":          params.Email,
				"Nickname":       params.Nickname,
				"Birthdate":      params.Birthdate,
			}))
		} else {
			returnTo := url.QueryEscape("/settings/profile")
			slog.Info(returnTo)
			redirectTo := fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, returnTo)
			slog.Info(redirectTo)
			redirect(c, redirectTo)
		}
		return
	}

	// Settings Flow の送信(完了)
	output, err := p.d.Kratos.UpdateSettingsFlowProfile(kratos.UpdateSettingsFlowProfileInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Traits: map[string]interface{}{
			"email":     params.Email,
			"nickname":  params.Nickname,
			"birthdate": params.Birthdate,
		},
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/_form.html", viewParameters(c, gin.H{
			"Title":         "Settings",
			"CsrfToken":     csrfToken,
			"ErrorMessages": output.ErrorMessages,
			"Email":         params.Email,
			"Nickname":      params.Nickname,
			"Birthdate":     params.Birthdate,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Settings flow成功時はVerification flowへリダイレクト
	// c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	c.Writer.Header().Set("HX-Location", fmt.Sprintf("%s/", generalEndpoint))
	c.Status(200)
}

func (p *Provider) updateProfile(c *gin.Context, params settingsProfileEditViewParams) error {
	session := getSession(c)
	params = mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{
		Email:     params.Email,
		Nickname:  params.Nickname,
		Birthdate: params.Birthdate,
	}, session)
	slog.Info("params", "email", params.Email, "nickname", params.Nickname, "birthdate", params.Birthdate)

	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: params.FlowID,
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// Settings Flow の送信(完了)
	updateOutput, err := p.d.Kratos.UpdateSettingsFlowProfile(kratos.UpdateSettingsFlowProfileInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    output.FlowID,
		CsrfToken: output.CsrfToken,
		Traits: map[string]interface{}{
			"email":     params.Email,
			"nickname":  params.Nickname,
			"birthdate": params.Birthdate,
		},
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, updateOutput.Cookies)

	return nil
}

// Home画面（ログイン必須）レンダリング
type item struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

var items = []item{
	{
		Name:        "Item1",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item2",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
}

func (p *Provider) handleGetHome(c *gin.Context) {
	c.HTML(http.StatusOK, "home/index.html", viewParameters(c, gin.H{
		"Title": "Home",
		"Items": items,
	}))
}

func (p *Provider) handleGetItemDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "item/detail.html", viewParameters(c, gin.H{
		"Title": "Home",
		"Items": items,
	}))
}
