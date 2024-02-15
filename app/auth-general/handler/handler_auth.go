package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"net/url"
)

// ------------------------- Authentication Registration -------------------------

// Handler GET /auth/registration
type handleGetAuthRegistrationdRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetAuthRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	reqParams := handleGetAuthRegistrationdRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Registration Flow の作成 or 取得
	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRegistrationFlow(kratos.CreateOrGetRegistrationFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthRegistrationIndex, viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s?flow=%s", routePaths.AuthRegistration, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, templatePaths.AuthRegistrationIndex, viewParameters(session, r, map[string]any{
		"RegistrationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
	}))
}

// Handler POST /auth/registration
type handlePostAuthRegistrationRequestParams struct {
	FlowID               string `validate:"required,uuid4"`
	CsrfToken            string `validate:"required"`
	Email                string `validate:"required,email" ja:"メールアドレス"`
	Password             string `validate:"required" ja:"パスワード"`
	PasswordConfirmation string `validate:"required" ja:"パスワード確認"`
}

func (p *handlePostAuthRegistrationRequestParams) validate() map[string]string {
	err := validate.Struct(p)
	if err != nil {
		slog.Error(err.Error())
	}
	fieldErrors := validationFieldErrors(validate.Struct(p))
	if p.Password != p.PasswordConfirmation {
		fieldErrors["Password"] = "パスワードとパスワード確認が一致しません"
	}
	return fieldErrors
}

func (p *Provider) handlePostAuthRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// リクエストパラメータのバリデーション
	reqParams := handlePostAuthRegistrationRequestParams{
		FlowID:               r.URL.Query().Get("flow"),
		CsrfToken:            r.PostFormValue("csrf_token"),
		Email:                r.PostFormValue("email"),
		Password:             r.PostFormValue("password"),
		PasswordConfirmation: r.PostFormValue("password-confirmation"),
	}
	// validationFieldErrors := reqParams.validate()
	// if len(validationFieldErrors) > 0 {
	// 	tmpl.ExecuteTemplate(w, templatePaths.AuthRegistration_Form, viewParameters(session, r, map[string]any{
	// 		"RegistrationFlowID":   reqParams.FlowID,
	// 		"CsrfToken":            reqParams.CsrfToken,
	// 		"Email":                reqParams.Email,
	// 		"Password":             reqParams.Password,
	// 		"ValidationFieldError": validationFieldErrors,
	// 	}))
	// 	return
	// }

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.FlowID,
		Email:     reqParams.Email,
		Password:  reqParams.Password,
		CsrfToken: reqParams.CsrfToken,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, templatePaths.AuthRegistration_Form, viewParameters(session, r, map[string]any{
			"RegistrationFlowID": reqParams.FlowID,
			"CsrfToken":          reqParams.CsrfToken,
			"Email":              reqParams.Email,
			"Password":           reqParams.Password,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Registration flow成功時はVerification flowへリダイレクト
	redirect(w, r, fmt.Sprintf("%s?flow=%s", routePaths.AuthRegistration, output.VerificationFlowID))
	w.WriteHeader(http.StatusOK)
}

// ------------------------- Authentication Verification -------------------------

// Handler GET /auth/verification handler
type handleGetAuthVerificationdRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetAuthVerification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	reqParams := handleGetAuthVerificationdRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerificationIndex, viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s?flow=%s", routePaths.AuthVerification, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, templatePaths.AuthVerificationIndex, viewParameters(session, r, map[string]any{
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler GET /auth/verification/code handler
type handleGetAuthVerificationCodeRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetAuthVerificationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	reqParams := handleGetAuthVerificationCodeRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Verification Flow の作成 or 取得
	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetVerificationFlow(kratos.CreateOrGetVerificationFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s?flow=%s", routePaths.AuthVerification, output.FlowID))
		return
	}

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
		"VerificationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"IsUsedFlow":         output.IsUsedFlow,
	}))
}

// Handler POST /auth/verification/email
type handlePostVerificationEmailRequestParams struct {
	flowID    string `validate:"uuid4"`
	csrfToken string `validate:"required"`
	email     string `validate:"required,email" ja:"メールアドレス"`
}

func (p *handlePostVerificationEmailRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostVerificationEmailRequestParams{
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		email:     r.PostFormValue("email"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
			"VerificationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Email":                reqParams.email,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Email:     reqParams.email,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
			"VerificationFlowID": reqParams.flowID,
			"CsrfToken":          reqParams.csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
		"VerificationFlowID": reqParams.flowID,
		"CsrfToken":          reqParams.csrfToken,
		"ErrorMessages":      output.ErrorMessages,
	}))
}

// Handler POST /auth/verification/code
type handlePostVerificationCodeRequestParams struct {
	flowID    string `validate:"uuid4"`
	csrfToken string `validate:"required"`
	code      string `validate:"required,len=6,number" ja:"検証コード"`
}

func (p *handlePostVerificationCodeRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostVerificationCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostVerificationCodeRequestParams{
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		code:      r.PostFormValue("code"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
			"VerificationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Code":                 reqParams.code,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		Code:      reqParams.code,
		CsrfToken: reqParams.csrfToken,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthVerification_CodeForm, viewParameters(session, r, map[string]any{
			"VerificationFlowID": reqParams.flowID,
			"CsrfToken":          reqParams.csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Loign 画面へリダイレクト
	redirect(w, r, routePaths.AuthLogin)
}

// ------------------------- Authentication Login -------------------------

// Handler GET /auth/login
type handleGetAuthLoginRequestParams struct {
	cookie string
	flowID string
}

// func (p *Provider) handleGetAuthLogin(w http.ResponseWriter, r *http.Request) {
func (p *Provider) handleGetAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	reqParams := handleGetAuthLoginRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// 認証済みの場合、認証時刻の更新を実施
	// プロフィール設定時に認証時刻が一定期間内である必要があり、過ぎている場合はログイン画面へリダイレクトし、ログインを促している
	refresh := isAuthenticated(session)

	// Login Flow の作成 or 取得
	output, err := p.d.Kratos.CreateOrGetLoginFlow(kratos.CreateOrGetLoginFlowInput{
		Cookie:  reqParams.cookie,
		FlowID:  reqParams.flowID,
		Refresh: refresh,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, templatePaths.AuthLoginIndex, viewParameters(session, r, map[string]any{
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
			redirectTo = fmt.Sprintf("%s?flow=%s", routePaths.AuthLogin, output.FlowID)
		} else {
			redirectTo = fmt.Sprintf("%s?flow=%s&return_to=%s", routePaths.AuthLogin, output.FlowID, returnTo)
		}
		redirect(w, r, redirectTo)
		return
	}

	var information string
	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィール更新のために、再度ログインをお願いします。"
	}

	w.WriteHeader(http.StatusOK)
	tmpl.ExecuteTemplate(w, templatePaths.AuthLoginIndex, viewParameters(session, r, map[string]any{
		"LoginFlowID": output.FlowID,
		"ReturnTo":    returnTo,
		"Information": information,
		"CsrfToken":   output.CsrfToken,
	}))
}

// Handler POST /auth/login
type handlePostAuthLoginRequestParams struct {
	flowID     string `validate:"uuid4"`
	csrfToken  string `validate:"required"`
	identifier string `validate:"required,email" ja:"メールアドレス"`
	password   string `validate:"required" ja:"パスワード"`
}

func (p *handlePostAuthLoginRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostAuthLoginRequestParams{
		flowID:     r.URL.Query().Get("flow"),
		csrfToken:  r.PostFormValue("csrf_token"),
		identifier: r.PostFormValue("identifier"),
		password:   r.PostFormValue("password"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, templatePaths.AuthLogin_Form, viewParameters(session, r, map[string]any{
			"LoginFlowID":          reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Identifier":           reqParams.identifier,
			"Password":             reqParams.password,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Login Flow 更新
	output, err := p.d.Kratos.UpdateLoginFlow(kratos.UpdateLoginFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Identifier: reqParams.identifier,
		Password:   reqParams.password,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthLogin_Form, viewParameters(session, r, map[string]any{
			"LoginFlowID":   reqParams.flowID,
			"CsrfToken":     reqParams.csrfToken,
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
		err := p.updateProfile(w, r, updateProfileParams{
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
		redirectTo = routePaths.Top
	}
	redirect(w, r, redirectTo)
}

// ------------------------- Authentication Logout -------------------------

// Handler POST /auth/logout
type handlePostAuthLogoutRequestParams struct {
	cookie string
}

func (p *Provider) handlePostAuthLogout(w http.ResponseWriter, r *http.Request) {
	reqParams := handlePostAuthLogoutRequestParams{
		cookie: r.Header.Get("Cookie"),
	}
	// Logout
	_, err := p.d.Kratos.Logout(kratos.LogoutFlowInput{
		Cookie: reqParams.cookie,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     cookieParams.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     cookieParams.Path,
		Domain:   cookieParams.Domain,
		Secure:   cookieParams.Secure,
		HttpOnly: true,
	})
	if err != nil {
		redirect(w, r, routePaths.Top)
		w.WriteHeader(http.StatusOK)
		return
	}

	redirect(w, r, routePaths.Top)
	w.WriteHeader(http.StatusOK)
}

// ------------------------- Authentication Recovery -------------------------

// Handler GET /auth/recovery
type handleGetAuthRecoveryRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetAuthRecovery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	reqParams := handleGetAuthRecoveryRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Recovery Flow の作成 or 取得
	// Recovery flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetRecoveryFlow(kratos.CreateOrGetRecoveryFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, templatePaths.AuthRecoveryIndex, viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if output.IsNewFlow {
		redirect(w, r, fmt.Sprintf("%s?flow=%s", routePaths.AuthRecovery, output.FlowID))
		return
	}

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, templatePaths.AuthRecoveryIndex, viewParameters(session, r, map[string]any{
		"RecoveryFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
	}))
}

// Handler POST /recovery/email
type handlePostAuthRecoveryEmailRequestParams struct {
	flowID    string `validate:"uuid4"`
	csrfToken string `validate:"required"`
	email     string `validate:"required,email" ja:"メールアドレス"`
}

func (p *handlePostAuthRecoveryEmailRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostAuthRecoveryEmailRequestParams{
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		email:     r.PostFormValue("email"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, templatePaths.AuthRecovery_CodeForm, viewParameters(session, r, map[string]any{
			"RecoveryFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Email":                reqParams.email,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Email:     reqParams.email,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, templatePaths.AuthRecovery_CodeForm, viewParameters(session, r, map[string]any{
			"RecoveryFlowID": reqParams.flowID,
			"CsrfToken":      reqParams.csrfToken,
			"Email":          reqParams.email,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, templatePaths.AuthRecovery_CodeForm, viewParameters(session, r, map[string]any{
		"RecoveryFlowID":           reqParams.flowID,
		"CsrfToken":                reqParams.csrfToken,
		"Email":                    reqParams.email,
		"ShowRecoveryAnnouncement": true,
	}))
}

// Handler POST /recovery/code
type handlePostAuthRecoveryCodeRequestParams struct {
	flowID    string `validate:"uuid4"`
	csrfToken string `validate:"required"`
	code      string `validate:"required,,len=6,number" ja:"復旧コード"`
}

func (p *handlePostAuthRecoveryCodeRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthRecoveryCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostAuthRecoveryCodeRequestParams{
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		code:      r.PostFormValue("code"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, templatePaths.AuthRecovery_CodeForm, viewParameters(session, r, map[string]any{
			"RecoveryFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Code":                 reqParams.code,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Code:      reqParams.code,
	})
	if err != nil && output.RedirectBrowserTo == "" {
		tmpl.ExecuteTemplate(w, templatePaths.AuthRecovery_CodeForm, viewParameters(session, r, map[string]any{
			"RecoveryFlowID": reqParams.flowID,
			"CsrfToken":      reqParams.csrfToken,
			"Code":           reqParams.code,
			"ErrorMessages":  output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)
	redirect(w, r, fmt.Sprintf("%s&from=recovery", output.RedirectBrowserTo))
	w.WriteHeader(http.StatusOK)
}
