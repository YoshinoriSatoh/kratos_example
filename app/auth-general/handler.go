package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func getKratosSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		output, err := ToSession(ToSessionInput{
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

func redirectIfExistsTraitsFieldsNotFilledIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		targetPaths := []string{"/settings/profile/edit", "/login"}
		if IsAuthenticated(c) &&
			existsTraitsFieldsNotFilledIn(session) &&
			!slices.Contains(targetPaths, c.Request.URL.Path) {
			c.Redirect(303, fmt.Sprintf("%s/settings/profile/edit", generalEndpoint))
		} else {
			c.Next()
		}
	}
}

func requireAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAuthenticated(c) {
			slog.Info("Authenticated")
			c.Next()
		} else {
			slog.Info("Not Authenticated")
			// c.AbortWithStatus(http.StatusUnauthorized)
			c.Redirect(303, fmt.Sprintf("%s/error/unauthorized", generalEndpoint))
		}
	}
}

func setCookieToResponseHeader(c *gin.Context, cookies []string) {
	for _, cookie := range cookies {
		c.Writer.Header().Add("Set-Cookie", cookie)
	}
}

// Handler GET /registration
func handleGetRegistration(c *gin.Context) {
	// Registration Flow の作成 or 取得
	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetRegistrationFlow(CreateOrGetRegistrationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "registration/index.html", ViewParameters(c, gin.H{
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
	c.HTML(http.StatusOK, "registration/index.html", ViewParameters(c, gin.H{
		"Title":              "Registration",
		"RegistrationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
	}))
}

// Handler POST /registration
func handlePostRegistration(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Registration Flow 更新
	output, err := UpdateRegistrationFlow(UpdateRegistrationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     c.PostForm("email"),
		Password:  c.PostForm("password"),
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "registration/_form.html", ViewParameters(c, gin.H{
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
func handleGetVerification(c *gin.Context) {
	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetVerificationFlow(CreateOrGetVerificationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/index.html", ViewParameters(c, gin.H{
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
	c.HTML(http.StatusOK, "verification/index.html", ViewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler GET /verification/code handler
func handleGetVerificationCode(c *gin.Context) {
	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetVerificationFlow(CreateOrGetVerificationFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/code.html", ViewParameters(c, gin.H{
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
	c.HTML(http.StatusOK, "verification/code.html", ViewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler POST /verification/email
func handlePostVerificationEmail(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	email := c.PostForm("email")

	// Verification Flow 更新
	output, err := UpdateVerificationFlow(UpdateVerificationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     email,
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/_code_form.html", ViewParameters(c, gin.H{
			"Title":              "Verification",
			"VerificationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	c.HTML(http.StatusOK, "verification/_code_form.html", ViewParameters(c, gin.H{
		"Title":              "Verification",
		"VerificationFlowID": flowID,
		"CsrfToken":          csrfToken,
		"ErrorMessages":      output.ErrorMessages,
	}))
}

// Handler POST /verification/code
func handlePostVerificationCode(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	code := c.PostForm("code")

	// Verification Flow 更新
	output, err := UpdateVerificationFlow(UpdateVerificationFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		Code:      code,
		CsrfToken: csrfToken,
	})
	if err != nil {
		c.HTML(http.StatusOK, "verification/_code_form.html", ViewParameters(c, gin.H{
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
func handleGetLogin(c *gin.Context) {
	var refresh bool

	// 認証済みの場合、認証時刻の更新を実施
	// プロフィール設定時に認証時刻が一定期間内である必要があり、過ぎている場合はログイン画面へリダイレクトし、ログインを促している
	if IsAuthenticated(c) {
		refresh = true
	}

	// Login Flow の作成 or 取得
	output, err := CreateOrGetLoginFlow(CreateOrGetLoginFlowInput{
		Cookie:  c.Request.Header.Get("Cookie"),
		FlowID:  c.Query("flow"),
		Refresh: refresh,
	})
	if err != nil {
		c.HTML(http.StatusOK, "login/index.html", ViewParameters(c, gin.H{
			"Title":         "Login",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/login?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// Login flowを新規作成した場合は、FlowIDを含めてリダイレクト
	// return_to
	//   指定時: ログイン後にreturn_toで指定されたURLへリダイレクト
	//   未指定時: ログイン後にホーム画面へリダイレクト
	returnTo := c.Query("return_to")
	if output.IsNewFlow {
		c.Redirect(303, fmt.Sprintf("%s/login?flow=%s&return_to=%s", generalEndpoint, output.FlowID, returnTo))
		return
	}

	c.HTML(http.StatusOK, "login/index.html", ViewParameters(c, gin.H{
		"Title":       "Login",
		"LoginFlowID": output.FlowID,
		"ReturnTo":    returnTo,
		"CsrfToken":   output.CsrfToken,
	}))
}

// Handler POST /login
func handlePostLogin(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Login Flow 更新
	output, err := UpdateLoginFlow(UpdateLoginFlowInput{
		Cookie:     c.Request.Header.Get("Cookie"),
		FlowID:     flowID,
		CsrfToken:  csrfToken,
		Identifier: c.PostForm("identifier"),
		Password:   c.PostForm("password"),
	})
	if err != nil {
		c.Status(400)
		c.HTML(http.StatusOK, "login/_form.html", ViewParameters(c, gin.H{
			"Title":         "Login",
			"LoginFlowID":   flowID,
			"CsrfToken":     csrfToken,
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// Login flow成功時:
	//   Traitsに設定させたい項目がまだ未設定の場合、Settings(profile)へリダイレクト
	//   はホーム画面へリダイレクト
	returnTo := c.Query("return_to")
	if returnTo != "" {
		c.Writer.Header().Set("HX-Redirect", returnTo)
	} else {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	}
	c.Status(200)
}

// Handler POST /logout
func handlePostLogout(c *gin.Context) {
	// Logout
	_, err := Logout(LogoutFlowInput{
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
func handleGetRecovery(c *gin.Context) {
	// Recovery Flow の作成 or 取得
	// Recovery flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetRecoveryFlow(CreateOrGetRecoveryFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "recovery/index.html", ViewParameters(c, gin.H{
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
	c.HTML(http.StatusOK, "recovery/index.html", ViewParameters(c, gin.H{
		"Title":          "Recovery",
		"RecoveryFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
	}))
}

// Handler POST /recovery/email
func handlePostRecoveryEmail(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Recovery Flow 更新
	output, err := UpdateRecoveryFlow(UpdateRecoveryFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Email:     c.PostForm("email"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "recovery/_code_form.html", ViewParameters(c, gin.H{
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
	c.HTML(http.StatusOK, "recovery/_code_form.html", ViewParameters(c, gin.H{
		"Title":                    "Recovery",
		"RecoveryFlowID":           flowID,
		"CsrfToken":                csrfToken,
		"ShowRecoveryAnnouncement": true,
	}))
}

// Handler POST /recovery/code
func handlePostRecoveryCode(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Recovery Flow 更新
	output, err := UpdateRecoveryFlow(UpdateRecoveryFlowInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Code:      c.PostForm("code"),
	})
	if err != nil && output.RedirectBrowserTo == "" {
		c.HTML(http.StatusOK, "recovery/_code_form.html", ViewParameters(c, gin.H{
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
func handleGetPasswordSettings(c *gin.Context) {
	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetSettingsFlow(CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})

	if err != nil {
		c.HTML(http.StatusOK, "settings/password/index.html", ViewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// flowの情報に従ってレンダリング
	c.HTML(http.StatusOK, "settings/password/index.html", ViewParameters(c, gin.H{
		"Title":                "Settings",
		"SettingsFlowID":       output.FlowID,
		"CsrfToken":            output.CsrfToken,
		"RedirectFromRecovery": c.Query("from") == "recovery",
	}))
}

// Handler POST /settings/password
func handlePostSettingsPassword(c *gin.Context) {
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")

	// Setting Flow 更新
	output, err := UpdateSettingsFlowPassword(UpdateSettingsFlowPasswordInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Password:  c.PostForm("password"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/password/_form.html", ViewParameters(c, gin.H{
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
func handleGetSettingsProfile(c *gin.Context) {
	session := GetSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetSettingsFlow(CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/index.html", ViewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// flowの情報に従ってレンダリング
	nickname, ok := session.Identity.Traits.(map[string]interface{})["nickname"].(string)
	if !ok {
		nickname = ""
	}
	birthdate, ok := session.Identity.Traits.(map[string]interface{})["birthdate"].(string)
	if !ok {
		birthdate = ""
	}
	c.HTML(http.StatusOK, "settings/profile/index.html", ViewParameters(c, gin.H{
		"Title":          "Profile",
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Nickname":       nickname,
		"Birthdate":      birthdate,
	}))
}

// Handler GET /settings/profile/edit
func handleGetSettingsProfileEdit(c *gin.Context) {
	session := GetSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetSettingsFlow(CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/edit.html", ViewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if NeedLoginWhenPrivilegedAccess(c) {
		redirectTo := fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, "/settings/profile/edit?flow="+output.FlowID)
		if c.Request.Header.Get("HX-Request") == "true" {
			c.Writer.Header().Set("HX-Redirect", redirectTo)
		} else {
			c.Redirect(303, redirectTo)
		}
		c.Status(200)
		return
	}

	// プロフィール更新時にセッションがprivileged_session_max_age を過ぎていてログイン画面にログインされ、ログイン後に本APIへリダイレクトされた場合、
	// 編集画面に入力したフォームの値がCookieに保存されているため、それを取得してフォームに表示する
	// Cookieは取得後に削除する
	params, err := loadSettingsProfileEditParamsFromCookie(c)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// クッキーに編集画面のパラメータが保存されていない場合は、セッションから現在の値を取得
	traits := session.Identity.Traits.(map[string]interface{})
	slog.Info(fmt.Sprintf("%v", traits))
	if params.Nickname == "" {
		params.Nickname = getValueFromTraits(traits, "nickname")
	}
	if params.Birthdate == "" {
		params.Nickname = getValueFromTraits(traits, "birthdate")
	}

	var information string
	if existsTraitsFieldsNotFilledIn(session) {
		information = "プロフィールの入力をお願いします"
	}

	c.HTML(http.StatusOK, "settings/profile/edit.html", ViewParameters(c, gin.H{
		"Title":          "Profile",
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Nickname":       params.Nickname,
		"Birthdate":      params.Birthdate,
		"Infomation":     information,
	}))
}

// Handler GET /settings/profile/_form
func handleGetSettingsProfileForm(c *gin.Context) {
	session := GetSession(c)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := CreateOrGetSettingsFlow(CreateOrGetSettingsFlowInput{
		Cookie: c.Request.Header.Get("Cookie"),
		FlowID: c.Query("flow"),
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/_form.html", ViewParameters(c, gin.H{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(c, output.Cookies)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if NeedLoginWhenPrivilegedAccess(c) {
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, "/settings/profile/edit?flow="+output.FlowID))
		c.Status(200)
		return
	}

	// セッションから現在の値を取得
	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{}, session)

	c.HTML(http.StatusOK, "settings/profile/_form.html", ViewParameters(c, gin.H{
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
	Nickname  string
	Birthdate string
}

func (f handlePostSettingsProfileRequestPostForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Nickname, validation.Required, validation.Length(5, 20)),
		validation.Field(&f.Birthdate, validation.Required, validation.Date("2006-01-02")),
	)
}

func handlePostSettingsProfile(c *gin.Context) {
	session := GetSession(c)
	flowID := c.Query("flow")
	csrfToken := c.PostForm("csrf_token")
	email := c.PostForm("email")
	requestPostForm := handlePostSettingsProfileRequestPostForm{
		Nickname:  c.PostForm("nickname"),
		Birthdate: c.PostForm("birthdate"),
	}

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
		c.HTML(http.StatusOK, "settings/profile/_form.html", ViewParameters(c, gin.H{
			"SettingsFlowID":  flowID,
			"CsrfToken":       csrfToken,
			"ErrorMessages":   []string{"Error"},
			"Email":           email,
			"Nickname":        requestPostForm.Nickname,
			"Birthdate":       requestPostForm.Nickname,
			"ValidationError": validationErrors,
		}))
		return
	}

	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{
		Email:     email,
		Nickname:  requestPostForm.Nickname,
		Birthdate: requestPostForm.Birthdate,
	}, session)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if NeedLoginWhenPrivilegedAccess(c) {
		err := saveSettingsProfileEditParamsToCookie(c, params)
		if err != nil {
			c.HTML(http.StatusOK, "settings/profile/_form.html", ViewParameters(c, gin.H{
				"Title":          "Settings",
				"SettingsFlowID": flowID,
				"CsrfToken":      csrfToken,
				"ErrorMessages":  []string{"Error"},
				"Email":          params.Email,
				"Nickname":       params.Nickname,
				"Birthdate":      params.Birthdate,
			}))
		}
		c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, "/settings/profile/edit?flow="+flowID))
		c.Status(200)
		return
	}

	// Settings Flow の送信(完了)
	output, err := UpdateSettingsFlowProfile(UpdateSettingsFlowProfileInput{
		Cookie:    c.Request.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Email:     params.Email,
		Nickname:  params.Nickname,
		Birthdate: params.Birthdate,
	})
	if err != nil {
		c.HTML(http.StatusOK, "settings/profile/_form.html", ViewParameters(c, gin.H{
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
	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("%s/", generalEndpoint))
	c.Status(200)
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

func handleGetHome(c *gin.Context) {
	c.HTML(http.StatusOK, "home/index.html", ViewParameters(c, gin.H{
		"Title": "Home",
		"Items": items,
	}))
}

func handleGetItemDetail(c *gin.Context) {
	c.HTML(http.StatusOK, "item/detail.html", ViewParameters(c, gin.H{
		"Title": "Home",
		"Items": items,
	}))
}
