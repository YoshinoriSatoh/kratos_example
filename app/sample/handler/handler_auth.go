package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"
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

	slog.Info(fmt.Sprintf("%v", reqParams))

	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateRegistrationFlow(kratos.CreateRegistrationFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/registration", output.FlowID))
		return
	}

	// Registration Flow の 取得
	output, err := p.d.Kratos.GetRegistrationFlow(kratos.GetRegistrationFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	if output.RequestFromOidc {
		adminListIdentitiesOutput, err := p.d.Kratos.AdminListIdentities(kratos.AdminListIdentitiesInput{
			CredentialIdentifier: output.Traits.Email,
			Cookie:               reqParams.cookie,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}

		slog.Info(fmt.Sprintf("%v", adminListIdentitiesOutput))

		if len(adminListIdentitiesOutput.Identities) > 0 {
			updateRegistrationOutput, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
				Cookie:     r.Header.Get("Cookie"),
				RemoteAddr: r.RemoteAddr,
				FlowID:     reqParams.flowID,
				CsrfToken:  output.CsrfToken,
				Method:     "oidc",
				Provider:   "google",
				Traits:     adminListIdentitiesOutput.Identities[0].Traits,
			})
			if err != nil || len(output.ErrorMessages) > 0 {
				pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form.html", viewParameters(session, r, map[string]any{
					"RegistrationFlowID": reqParams.flowID,
					"CsrfToken":          output.CsrfToken,
					"Traits":             adminListIdentitiesOutput.Identities[0].Traits,
					"ErrorMessages":      output.ErrorMessages,
				}))
				return
			}

			slog.Info(fmt.Sprintf("%v", updateRegistrationOutput))
			if updateRegistrationOutput.RedirectBrowserTo != "" {
				setCookieToResponseHeader(w, updateRegistrationOutput.Cookies)
				redirect(w, r, updateRegistrationOutput.RedirectBrowserTo)
			}
		}
	}

	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	w.WriteHeader(http.StatusOK)
	if output.RenderingType == kratos.RegistrationRenderingTypeOidc {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/oidc.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID": output.FlowID,
			"CsrfToken":          output.CsrfToken,
			"Traits":             output.Traits,
		}))
	} else {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/index.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID": output.FlowID,
			"CsrfToken":          output.CsrfToken,
		}))
	}
}

// Handler GET /auth/registration/passkey
type handleGetAuthRegistrationdPasskeyRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetAuthRegistrationPasskey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetAuthRegistrationdPasskeyRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	slog.Info(fmt.Sprintf("%v", reqParams))

	// Registration flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateRegistrationFlow(kratos.CreateRegistrationFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/passkey.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/registration/passkey", output.FlowID))
		return
	}

	// Registration Flow の 取得
	output, err := p.d.Kratos.GetRegistrationFlow(kratos.GetRegistrationFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/passkey.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	w.WriteHeader(http.StatusOK)
	pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/passkey.html", viewParameters(session, r, map[string]any{
		"RegistrationFlowID": output.FlowID,
		"CsrfToken":          output.CsrfToken,
		"Traits":             output.Traits,
		"PasskeyCreateData":  output.PasskeyCreateData,
	}))
}

// Handler POST /auth/registration
type handlePostAuthRegistrationRequestParams struct {
	FlowID               string        `validate:"required,uuid4"`
	CsrfToken            string        `validate:"required"`
	Traits               kratos.Traits `validate:"required"`
	Password             string        `validate:"required" ja:"パスワード"`
	PasswordConfirmation string        `validate:"required" ja:"パスワード確認"`
}

func (p *handlePostAuthRegistrationRequestParams) validate() map[string]string {
	err := pkgVars.validate.Struct(p)
	if err != nil {
		slog.Error(err.Error())
	}
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	if p.Password != p.PasswordConfirmation {
		fieldErrors["Password"] = "パスワードとパスワード確認が一致しません"
	}
	return fieldErrors
}

func (p *Provider) handlePostAuthRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// リクエストパラメータのバリデーション
	traits := kratos.Traits{
		Email:     r.PostFormValue("traits.email"),
		Firstname: r.PostFormValue("traits.firstname"),
		Lastname:  r.PostFormValue("traits.lastname"),
		Nickname:  r.PostFormValue("traits.nickname"),
	}
	traits.Birthdate, _ = time.Parse(pkgVars.birthdateFormat, r.PostFormValue("traits.birthdate"))
	reqParams := handlePostAuthRegistrationRequestParams{
		FlowID:               r.URL.Query().Get("flow"),
		CsrfToken:            r.PostFormValue("csrf_token"),
		Traits:               traits,
		Password:             r.PostFormValue("password"),
		PasswordConfirmation: r.PostFormValue("password-confirmation"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID":   reqParams.FlowID,
			"CsrfToken":            reqParams.CsrfToken,
			"Traits":               traits,
			"Password":             reqParams.Password,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.FlowID,
		CsrfToken:  reqParams.CsrfToken,
		Method:     "password",
		Traits:     traits,
		Password:   reqParams.Password,
	})
	if err != nil || len(output.ErrorMessages) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID": reqParams.FlowID,
			"CsrfToken":          reqParams.CsrfToken,
			"Traits":             traits,
			"Password":           reqParams.Password,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Registration flow成功時はVerification flowへリダイレクト
	redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/verification/code", output.VerificationFlowID))
	w.WriteHeader(http.StatusOK)
}

// Handler POST /auth/registration/oidc
type handlePostAuthRegistrationOidcRequestParams struct {
	FlowID    string        `validate:"required,uuid4"`
	CsrfToken string        `validate:"required"`
	Provider  string        `validate:"required"`
	Traits    kratos.Traits `validate:"required"`
}

func (p *handlePostAuthRegistrationOidcRequestParams) validate() map[string]string {
	err := pkgVars.validate.Struct(p)
	if err != nil {
		slog.Error(err.Error())
	}
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthRegistrationOidc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// リクエストパラメータのバリデーション
	traits := kratos.Traits{
		Email:     r.PostFormValue("traits.email"),
		Firstname: r.PostFormValue("traits.firstname"),
		Lastname:  r.PostFormValue("traits.lastname"),
		Nickname:  r.PostFormValue("traits.nickname"),
	}
	traits.Birthdate, _ = time.Parse(pkgVars.birthdateFormat, r.PostFormValue("traits.birthdate"))
	reqParams := handlePostAuthRegistrationOidcRequestParams{
		FlowID:    r.URL.Query().Get("flow"),
		CsrfToken: r.PostFormValue("csrf_token"),
		Provider:  r.PostFormValue("provider"),
		Traits:    traits,
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form_oidc.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID":   reqParams.FlowID,
			"CsrfToken":            reqParams.CsrfToken,
			"Traits":               traits,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.FlowID,
		CsrfToken:  reqParams.CsrfToken,
		Method:     "oidc",
		Provider:   reqParams.Provider,
		Traits:     traits,
	})
	if err != nil && output.RedirectBrowserTo == "" {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form_oidc.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID": reqParams.FlowID,
			"CsrfToken":          reqParams.CsrfToken,
			"Traits":             traits,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	redirect(w, r, output.RedirectBrowserTo)
	w.WriteHeader(http.StatusOK)
}

// Handler POST /auth/registration/passkey
type handlePostAuthRegistrationPasskeyRequestParams struct {
	FlowID          string        `validate:"required,uuid4"`
	CsrfToken       string        `validate:"required"`
	Traits          kratos.Traits `validate:"required"`
	PasskeyRegister string
}

func (p *handlePostAuthRegistrationPasskeyRequestParams) validate() map[string]string {
	err := pkgVars.validate.Struct(p)
	if err != nil {
		slog.Error(err.Error())
	}
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthRegistrationPasskey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	// リクエストパラメータのバリデーション
	traits := kratos.Traits{
		Email:     r.PostFormValue("traits.email"),
		Firstname: r.PostFormValue("traits.firstname"),
		Lastname:  r.PostFormValue("traits.lastname"),
		Nickname:  r.PostFormValue("traits.nickname"),
	}
	traits.Birthdate, _ = time.Parse(pkgVars.birthdateFormat, r.PostFormValue("traits.birthdate"))
	reqParams := handlePostAuthRegistrationPasskeyRequestParams{
		FlowID:          r.URL.Query().Get("flow"),
		CsrfToken:       r.PostFormValue("csrf_token"),
		Traits:          traits,
		PasskeyRegister: r.PostFormValue("passkey_register"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form_passkey.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID":   reqParams.FlowID,
			"CsrfToken":            reqParams.CsrfToken,
			"Traits":               traits,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Registration Flow 更新
	output, err := p.d.Kratos.UpdateRegistrationFlow(kratos.UpdateRegistrationFlowInput{
		Cookie:          r.Header.Get("Cookie"),
		RemoteAddr:      r.RemoteAddr,
		FlowID:          reqParams.FlowID,
		CsrfToken:       reqParams.CsrfToken,
		Method:          "passkey",
		Traits:          traits,
		PasskeyRegister: reqParams.PasskeyRegister,
	})
	if err != nil || len(output.ErrorMessages) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/registration/_form_passkey.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID": reqParams.FlowID,
			"CsrfToken":          reqParams.CsrfToken,
			"Traits":             traits,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	redirect(w, r, output.RedirectBrowserTo)
	// Registration flow成功時はVerification flowへリダイレクト
	// redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/verification/code", output.VerificationFlowID))
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

	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateVerificationFlow(kratos.CreateVerificationFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/verification", output.FlowID))
		return
	}

	// Verification Flow の作成 or 取得
	output, err := p.d.Kratos.GetVerificationFlow(kratos.GetVerificationFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/index.html", viewParameters(session, r, map[string]any{
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

	// Verification flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateVerificationFlow(kratos.CreateVerificationFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/code.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/registration", output.FlowID))
		return
	}

	// Verification Flow の作成 or 取得
	output, err := p.d.Kratos.GetVerificationFlow(kratos.GetVerificationFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// 検証コード入力フォーム、もしくは既にVerification Flow が完了している旨のメッセージをレンダリング
	w.WriteHeader(http.StatusOK)
	pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/code.html", viewParameters(session, r, map[string]any{
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
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
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
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/_code_form.html", viewParameters(session, r, map[string]any{
			"VerificationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Email":                reqParams.email,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Email:      reqParams.email,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/_code_form.html", viewParameters(session, r, map[string]any{
			"VerificationFlowID": reqParams.flowID,
			"CsrfToken":          reqParams.csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	w.WriteHeader(http.StatusOK)
	pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/_code_form.html", viewParameters(session, r, map[string]any{
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
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
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
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/_code_form.html", viewParameters(session, r, map[string]any{
			"VerificationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Code":                 reqParams.code,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Verification Flow 更新
	output, err := p.d.Kratos.UpdateVerificationFlow(kratos.UpdateVerificationFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		Code:       reqParams.code,
		CsrfToken:  reqParams.csrfToken,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/verification/_code_form.html", viewParameters(session, r, map[string]any{
			"VerificationFlowID": reqParams.flowID,
			"CsrfToken":          reqParams.csrfToken,
			"ErrorMessages":      output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// Loign 画面へリダイレクト
	redirect(w, r, "/auth/login")
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

	returnTo := url.QueryEscape(r.URL.Query().Get("return_to"))

	// Login flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateLoginFlow(kratos.CreateLoginFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
			Refresh:    refresh,
		})
		if err != nil {
			pkgVars.tmpl.ExecuteTemplate(w, "auth/login/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}

		slog.Info(fmt.Sprintf("%v", output))
		// Login flowを新規作成した場合は、FlowIDを含めてリダイレクト
		// return_to
		//   指定時: ログイン後にreturn_toで指定されたURLへリダイレクト
		//   未指定時: ログイン後にホーム画面へリダイレクト
		var redirectTo string
		if returnTo == "" {
			redirectTo = fmt.Sprintf("%s?flow=%s", "/auth/login", output.FlowID)
		} else {
			redirectTo = fmt.Sprintf("%s?flow=%s&return_to=%s", "/auth/login", output.FlowID, returnTo)
		}
		redirect(w, r, redirectTo)
		return
	}

	// Login Flow の 取得
	output, err := p.d.Kratos.GetLoginFlow(kratos.GetLoginFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/login/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	var information string
	var traits kratos.Traits
	showSocialLogin := true
	if output.DuplicateIdentifier != "" {
		traits.Email = output.DuplicateIdentifier
		showSocialLogin = false
		information = "メールアドレスとパスワードで登録された既存のアカウントが存在します。パスワードを入力してログインすると、Googleのアカウントと連携されます。"
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィール更新のために、再度ログインをお願いします。"
	}

	slog.Info("ShowSocialLogin", showSocialLogin)

	w.WriteHeader(http.StatusOK)
	pkgVars.tmpl.ExecuteTemplate(w, "auth/login/index.html", viewParameters(session, r, map[string]any{
		"LoginFlowID":      output.FlowID,
		"ReturnTo":         returnTo,
		"Information":      information,
		"CsrfToken":        output.CsrfToken,
		"Traits":           traits,
		"ShowSocialLogin":  showSocialLogin,
		"PasskeyChallenge": output.PasskeyChallenge,
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
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
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
		slog.Info(fmt.Sprintf("%v", validationFieldErrors))
		pkgVars.tmpl.ExecuteTemplate(w, "auth/login/_form.html", viewParameters(session, r, map[string]any{
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
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Identifier: reqParams.identifier,
		Password:   reqParams.password,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/login/_form.html", viewParameters(session, r, map[string]any{
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
		birthdate, _ := time.Parse(pkgVars.birthdateFormat, hookParams["birthdate"].(string))
		err := p.updateProfile(w, r, updateProfileParams{
			FlowID:    hookParams["flow_id"].(string),
			Email:     hookParams["email"].(string),
			Nickname:  hookParams["nickname"].(string),
			Birthdate: birthdate,
		})
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}

	// return_to 指定時はreturn_toへリダイレクト
	returnTo := r.URL.Query().Get("return_to")
	slog.Info(returnTo)
	var redirectTo string
	if returnTo != "" {
		redirectTo = returnTo
	} else {
		redirectTo = "/"
	}
	redirect(w, r, redirectTo)
}

// Handler POST /auth/login/oidc
type handlePostAuthLoginOidcRequestParams struct {
	flowID    string `validate:"uuid4"`
	csrfToken string `validate:"required"`
	provider  string `validate:"required"`
}

func (p *handlePostAuthLoginOidcRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	return fieldErrors
}

func (p *Provider) handlePostAuthLoginOidc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostAuthLoginOidcRequestParams{
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		provider:  r.PostFormValue("provider"),
	}
	slog.Info(fmt.Sprintf("%v", reqParams))
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		slog.Info(fmt.Sprintf("%v", validationFieldErrors))
		pkgVars.tmpl.ExecuteTemplate(w, "auth/login/_form.html", viewParameters(session, r, map[string]any{
			"LoginFlowID":          reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Login Flow 更新
	output, err := p.d.Kratos.UpdateOidcLoginFlow(kratos.UpdateOidcLoginFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Provider:   reqParams.provider,
	})
	if err != nil && output.RedirectBrowserTo == "" {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/login/_form.html", viewParameters(session, r, map[string]any{
			"LoginFlowID":   reqParams.flowID,
			"CsrfToken":     reqParams.csrfToken,
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	log.Println("output", "RedirectBrowserTo", output.RedirectBrowserTo)
	redirect(w, r, output.RedirectBrowserTo)
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
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     pkgVars.cookieParams.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     pkgVars.cookieParams.Path,
		Domain:   pkgVars.cookieParams.Domain,
		Secure:   pkgVars.cookieParams.Secure,
		HttpOnly: true,
	})
	if err != nil {
		redirect(w, r, "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	redirect(w, r, "/")
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

	// Recovery flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateRecoveryFlow(kratos.CreateRecoveryFlowInput{
			Cookie:     reqParams.cookie,
			RemoteAddr: r.RemoteAddr,
			FlowID:     reqParams.flowID,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/auth/recovery", output.FlowID))
		return
	}

	// Recovery Flow の 取得
	output, err := p.d.Kratos.GetRecoveryFlow(kratos.GetRecoveryFlowInput{
		Cookie:     reqParams.cookie,
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/index.html", viewParameters(session, r, map[string]any{
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
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
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
		pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/_code_form.html", viewParameters(session, r, map[string]any{
			"RecoveryFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Email":                reqParams.email,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Email:      reqParams.email,
	})
	if err != nil {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/_code_form.html", viewParameters(session, r, map[string]any{
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
	pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/_code_form.html", viewParameters(session, r, map[string]any{
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
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
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
		pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/_code_form.html", viewParameters(session, r, map[string]any{
			"RecoveryFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Code":                 reqParams.code,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Recovery Flow 更新
	output, err := p.d.Kratos.UpdateRecoveryFlow(kratos.UpdateRecoveryFlowInput{
		Cookie:     r.Header.Get("Cookie"),
		RemoteAddr: r.RemoteAddr,
		FlowID:     reqParams.flowID,
		CsrfToken:  reqParams.csrfToken,
		Code:       reqParams.code,
	})
	if err != nil && output.RedirectBrowserTo == "" {
		pkgVars.tmpl.ExecuteTemplate(w, "auth/recovery/_code_form.html", viewParameters(session, r, map[string]any{
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
