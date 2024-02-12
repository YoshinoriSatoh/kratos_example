package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Handler GET /public/health
func (p *Provider) handleGetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Handler GET /registration
func (p *Provider) handleGetRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Registration Flow の作成 or 取得
	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRegistrationFlow(kratos.CreateOrGetRegistrationFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "registration/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Registration",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s/registration?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, "registration/index.html", viewParameters(session, r, map[string]any{
		"Title":              "Registration",
		"RegistrationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
	}))
}

// Handler POST /registration
func (p *Provider) handlePostRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     r.PostFormValue("email"),
		Password:  r.PostFormValue("password"),
		CsrfToken: csrfToken,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "registration/_form.html", viewParameters(session, r, map[string]any{
			"Title":              "Registration",
			"RegistrationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Registration flow成功時はVerification flowへリダイレクト
	redirect(w, r, fmt.Sprintf("%s/verification/code?flow=%s", generalEndpoint, output.VerificationFlowID))
	w.WriteHeader(http.StatusOK)
}

// Handler GET /verification handler
func (p *Provider) handleGetVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "verification/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Verification",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, "verification/index.html", viewParameters(session, r, map[string]any{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler GET /verification/code handler
func (p *Provider) handleGetVerificationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "verification/code.html", viewParameters(session, r, map[string]any{
			"Title":         "Verification",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s/verification?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, "verification/code.html", viewParameters(session, r, map[string]any{
		"Title":              "Verification",
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler POST /verification/email
func (p *Provider) handlePostVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")
	email := r.PostFormValue("email")

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		Email:     email,
		CsrfToken: csrfToken,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "verification/_code_form.html", viewParameters(session, r, map[string]any{
			"Title":              "Verification",
			"VerificationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, "verification/_code_form.html", viewParameters(session, r, map[string]any{
		"Title":              "Verification",
		"VerificationFlowID": flowID,
		"CsrfToken":          csrfToken,
		"ErrorMessages":      output.ErrorMessages,
	}))
}

// Handler POST /verification/code
func (p *Provider) handlePostVerificationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")
	code := r.PostFormValue("code")

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		Code:      code,
		CsrfToken: csrfToken,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "verification/_code_form.html", viewParameters(session, r, map[string]any{
			"Title":              "Verification",
			"VerificationFlowID": flowID,
			"CsrfToken":          csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Loign 画面へリダイレクト
	redirect(w, r, fmt.Sprintf("%s/login", generalEndpoint))
}

// Handler GET /login
// func (p *Provider) handleGetLogin(w http.ResponseWriter, r *http.Request) {
func (p *Provider) handleGetLogin(w http.ResponseWriter, r *http.Request) {
	var refresh bool

	ctx := r.Context()
	session := getSession(ctx)

	// 認証済みの場合、認証時刻の更新を実施
	// プロフィール設定時に認証時刻が一定期間内である必要があり、過ぎている場合はログイン画面へリダイレクトし、ログインを促している
	if isAuthenticated(session) {
		refresh = true
	}

	// Login Flow の作成 or 取得
	output, err := p.d.Kratos.CreateOrGetLoginFlow(kratos.CreateOrGetLoginFlowInput{
		Cookie:  r.Header.Get("Cookie"),
		FlowID:  r.URL.Query().Get("flow"),
		Refresh: refresh,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "login/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Login",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Login flowを新規作成した場合は、FlowIDを含めてリダイレクト
	// return_to
	//   指定時: ログイン後にreturn_toで指定されたURLへリダイレクト
	//   未指定時: ログイン後にホーム画面へリダイレクト
	returnTo := url.QueryEscape(r.URL.Query().Get("return_to"))
	if output.IsNewFlow {
		var redirectTo string
		if returnTo == "" {
			redirectTo = fmt.Sprintf("%s/login?flow=%s", generalEndpoint, output.FlowID)
		} else {
			redirectTo = fmt.Sprintf("%s/login?flow=%s&return_to=%s", generalEndpoint, output.FlowID, returnTo)
		}
		redirect(w, r, redirectTo)
		return
	}

	var information string
	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィール更新のために、再度ログインをお願いします。"
	}

	slog.Info(fmt.Sprintf("%v", tmpl))

	tmpl.ExecuteTemplate(w, "login/index.html", viewParameters(session, r, map[string]any{
		"Title":       "Login",
		"LoginFlowID": output.FlowID,
		"ReturnTo":    returnTo,
		"Information": information,
		"CsrfToken":   output.CsrfToken,
	}))
}

// Handler POST /login
func (p *Provider) handlePostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")

	// Login Flow 更新
	output, err := p.d.Kratos.UpdateLoginFlow(kratos.UpdateLoginFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		FlowID:     flowID,
		CsrfToken:  csrfToken,
		Identifier: r.PostFormValue("identifier"),
		Password:   r.PostFormValue("password"),
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "login/_form.html", viewParameters(session, r, map[string]any{
			"Title":         "Login",
			"LoginFlowID":   flowID,
			"CsrfToken":     csrfToken,
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// ログインフック実行
	hook, err := loadAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	if hook.Operation == AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE {
		hookParams, _ := hook.Params.(map[string]interface{})
		err := p.updateProfile(w, r, settingsProfileEditViewParams{
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
	returnTo := r.URL.Query().Get("return_to")
	slog.Info(returnTo)
	var redirectTo string
	if returnTo != "" {
		redirectTo = returnTo
	} else {
		redirectTo = fmt.Sprintf("%s/", generalEndpoint)
	}
	redirect(w, r, redirectTo)
}

// Handler POST /logout
func (p *Provider) handlePostLogout(w http.ResponseWriter, r *http.Request) {
	// Logout
	_, err := p.d.Kratos.Logout(kratos.LogoutFlowInput{
		Cookie: r.Header.Get("Cookie"),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "kratos_general_session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
	})
	if err != nil {
		redirect(w, r, fmt.Sprintf("%s/login", generalEndpoint))
		w.WriteHeader(http.StatusOK)
		return
	}

	redirect(w, r, fmt.Sprintf("%s/", generalEndpoint))
	w.WriteHeader(http.StatusOK)
}

// Handler GET /recovery
func (p *Provider) handleGetRecovery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Recovery Flow の作成 or 取得
	// Recovery flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRecoveryFlow(kratos.CreateOrGetRecoveryFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "recovery/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Recovery",
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s/recovery?flow=%s", generalEndpoint, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, "recovery/index.html", viewParameters(session, r, map[string]any{
		"Title":          "Recovery",
		"RecoveryFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
	}))
}

// Handler POST /recovery/email
func (p *Provider) handlePostRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Email:     r.PostFormValue("email"),
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "recovery/_code_form.html", viewParameters(session, r, map[string]any{
			"Title":          "Recovery",
			"RecoveryFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, "recovery/_code_form.html", viewParameters(session, r, map[string]any{
		"Title":                    "Recovery",
		"RecoveryFlowID":           flowID,
		"CsrfToken":                csrfToken,
		"ShowRecoveryAnnouncement": true,
	}))
}

// Handler POST /recovery/code
func (p *Provider) handlePostRecoveryCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Code:      r.PostFormValue("code"),
	})
	if err != nil && output.RedirectBrowserTo == "" {
		tmpl.ExecuteTemplate(w, "recovery/_code_form.html", viewParameters(session, r, map[string]any{
			"Title":          "Recovery",
			"RecoveryFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)
	redirect(w, r, fmt.Sprintf("%s&from=recovery", output.RedirectBrowserTo))
	w.WriteHeader(http.StatusOK)
}

// Handler GET /settings/password
func (p *Provider) handleGetPasswordSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})

	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/password/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, "settings/password/index.html", viewParameters(session, r, map[string]any{
		"Title":                "Settings",
		"SettingsFlowID":       output.FlowID,
		"CsrfToken":            output.CsrfToken,
		"RedirectFromRecovery": r.URL.Query().Get("from") == "recovery",
	}))
}

// Handler POST /settings/password
func (p *Provider) handlePostSettingsPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")

	// Setting Flow 更新
	output, err := p.d.Kratos.UpdateSettingsFlowPassword(kratos.UpdateSettingsFlowPasswordInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Password:  r.PostFormValue("password"),
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/password/_form.html", viewParameters(session, r, map[string]any{
			"Title":          "Settings",
			"SettingsFlowID": flowID,
			"CsrfToken":      csrfToken,
			"ErrorMessages":  output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Settings flow成功時はVerification flowへリダイレクト
	redirect(w, r, fmt.Sprintf("%s/login", generalEndpoint))
	w.WriteHeader(http.StatusOK)
}

// Handler GET /settings/profile
func (p *Provider) handleGetSettingsProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/profile/index.html", viewParameters(session, r, map[string]any{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	email := session.GetValueFromTraits("email")
	nickname := session.GetValueFromTraits("nickname")
	birthdate := session.GetValueFromTraits("birthdate")
	var information string
	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィールを更新しました。"
		deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	}
	tmpl.ExecuteTemplate(w, "settings/profile/index.html", viewParameters(session, r, map[string]any{
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
func (p *Provider) handleGetSettingsProfileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/profile/edit.html", viewParameters(session, r, map[string]any{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// if output.IsNewFlow {
	// 	redirect(w, r, fmt.Sprintf("%s/settings/profile/edit?flow=%s", generalEndpoint, output.FlowID))
	// 	return
	// }

	// セッションから現在の値を取得
	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{}, session)

	var information string
	if existsTraitsFieldsNotFilledIn(session) {
		information = "プロフィールの入力をお願いします"
	}

	tmpl.ExecuteTemplate(w, "settings/profile/edit.html", viewParameters(session, r, map[string]any{
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
func (p *Provider) handleGetSettingsProfileForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: r.URL.Query().Get("flow"),
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/profile/_form.html", viewParameters(session, r, map[string]any{
			"Title":         "Settings",
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// // セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	// if NeedLoginWhenPrivilegedAccess(c) {
	// 	returnTo := url.QueryEscape("/settings/profile/edit")
	// 	redirectTo := fmt.Sprintf("%s/login?return_to=%s", generalEndpoint, returnTo)
	// 	redirect(w, r, redirectTo)
	// 	return
	// }

	// セッションから現在の値を取得
	params := mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{}, session)

	tmpl.ExecuteTemplate(w, "settings/profile/_form.html", viewParameters(session, r, map[string]any{
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
func (p *Provider) handlePostSettingsProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	slog.Info("handlePostSettingsProfile")

	flowID := r.URL.Query().Get("flow")
	csrfToken := r.PostFormValue("csrf_token")
	requestPostForm := handlePostSettingsProfileRequestPostForm{
		Email:     r.PostFormValue("email"),
		Nickname:  r.PostFormValue("nickname"),
		Birthdate: r.PostFormValue("birthdate"),
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
		tmpl.ExecuteTemplate(w, "settings/profile/_form.html", viewParameters(session, r, map[string]any{
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

	deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if session.NeedLoginWhenPrivilegedAccess() {
		slog.Info(fmt.Sprintf("%v", params))
		// err := saveSettingsProfileEditParamsToCookie(c, params)
		err := saveAfterLoginHook(w, afterLoginHook{
			Operation: AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE,
			Params:    params,
		}, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
		if err != nil {
			tmpl.ExecuteTemplate(w, "settings/profile/_form.html", viewParameters(session, r, map[string]any{
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
			redirect(w, r, redirectTo)
		}
		return
	}

	// Settings Flow の送信(完了)
	output, err := p.d.Kratos.UpdateSettingsFlowProfile(kratos.UpdateSettingsFlowProfileInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    flowID,
		CsrfToken: csrfToken,
		Traits: map[string]interface{}{
			"email":     params.Email,
			"nickname":  params.Nickname,
			"birthdate": params.Birthdate,
		},
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "settings/profile/_form.html", viewParameters(session, r, map[string]any{
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
	setCookieToResponseHeader(w, output.Cookies)

	// Settings flow成功時はVerification flowへリダイレクト
	redirect(w, r, fmt.Sprintf("%s/", generalEndpoint))
	w.WriteHeader(http.StatusOK)
}

func (p *Provider) updateProfile(w http.ResponseWriter, r *http.Request, params settingsProfileEditViewParams) error {
	ctx := r.Context()
	session := getSession(ctx)

	params = mergeSettingsProfileEditViewParams(settingsProfileEditViewParams{
		Email:     params.Email,
		Nickname:  params.Nickname,
		Birthdate: params.Birthdate,
	}, session)
	slog.Info("params", "email", params.Email, "nickname", params.Nickname, "birthdate", params.Birthdate)

	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: params.FlowID,
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// Settings Flow の送信(完了)
	updateOutput, err := p.d.Kratos.UpdateSettingsFlowProfile(kratos.UpdateSettingsFlowProfileInput{
		Cookie:    r.Header.Get("Cookie"),
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
	setCookieToResponseHeader(w, updateOutput.Cookies)

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

func (p *Provider) handleGetHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	tmpl.ExecuteTemplate(w, "home/index.html", viewParameters(session, r, map[string]any{
		"Title": "Home",
		"Items": items,
	}))
}

func (p *Provider) handleGetItemDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	r.PathValue("id")

	tmpl.ExecuteTemplate(w, "item/detail.html", viewParameters(session, r, map[string]any{
		"Title": "Home",
		"Items": items,
	}))
}
