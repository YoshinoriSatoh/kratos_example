package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Handler GET /my/password
type handleGetMyPasswordRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetMyPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetMyPasswordRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateSettingsFlow(kratos.CreateSettingsFlowInput{
			Cookie: reqParams.cookie,
			FlowID: reqParams.flowID,
		})
		if err != nil {
			pkgVars.tmpl.ExecuteTemplate(w, "my/password/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/my/password", output.FlowID))
		return
	}

	// Setting Flow の作成 or 取得
	output, err := p.d.Kratos.GetSettingsFlow(kratos.GetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		pkgVars.tmpl.ExecuteTemplate(w, "my/password/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	pkgVars.tmpl.ExecuteTemplate(w, "my/password/index.html", viewParameters(session, r, map[string]any{
		"SettingsFlowID":       output.FlowID,
		"CsrfToken":            output.CsrfToken,
		"RedirectFromRecovery": reqParams.flowID == "recovery",
	}))
}

// Handler POST /my/password
type handlePostMyPasswordRequestParams struct {
	flowID               string `validate:"uuid4"`
	csrfToken            string `validate:"required"`
	password             string `validate:"required" ja:"パスワード"`
	passwordConfirmation string `validate:"required" ja:"パスワード確認"`
}

func (p *handlePostMyPasswordRequestParams) validate() map[string]string {
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	if p.password != p.passwordConfirmation {
		fieldErrors["Password"] = "パスワードとパスワード確認が一致しません"
	}
	return fieldErrors
}

func (p *Provider) handlePostMyPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)
	slog.Info("handlePostMyPassword")

	reqParams := handlePostMyPasswordRequestParams{
		flowID:               r.URL.Query().Get("flow"),
		csrfToken:            r.PostFormValue("csrf_token"),
		password:             r.PostFormValue("password"),
		passwordConfirmation: r.PostFormValue("password-confirmation"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		slog.Info(fmt.Sprintf("%v", validationFieldErrors))
		pkgVars.tmpl.ExecuteTemplate(w, "my/password/_form.html", viewParameters(session, r, map[string]any{
			"SettingsFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Password":             reqParams.password,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}
	slog.Info(fmt.Sprintf("%v", reqParams))

	// Setting Flow 更新
	output, err := p.d.Kratos.UpdateSettingsFlow(kratos.UpdateSettingsFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Method:    "password",
		Password:  reqParams.password,
	})
	if err != nil {
		slog.Info(err.Error())
		pkgVars.tmpl.ExecuteTemplate(w, "my/password/_form.html", viewParameters(session, r, map[string]any{
			"SettingsFlowID": reqParams.flowID,
			"CsrfToken":      reqParams.csrfToken,
			"Password":       reqParams.password,
			"ErrorMessages":  output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	redirect(w, r, "/")
	w.WriteHeader(http.StatusOK)
}

// Handler GET /my/profile
type handleGetMyProfileRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetMyProfileRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateSettingsFlow(kratos.CreateSettingsFlowInput{
			Cookie: reqParams.cookie,
			FlowID: reqParams.flowID,
		})
		if err != nil {
			pkgVars.tmpl.ExecuteTemplate(w, "my/profile/index.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/my/profile", output.FlowID))
		return
	}

	// Setting Flow の作成 or 取得
	output, err := p.d.Kratos.GetSettingsFlow(kratos.GetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		pkgVars.tmpl.ExecuteTemplate(w, "my/profile/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	var information string
	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィールを更新しました。"
		deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	}
	pkgVars.tmpl.ExecuteTemplate(w, "my/profile/index.html", viewParameters(session, r, map[string]any{
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          session.Identity.Traits.Email,
		"Firstname":      session.Identity.Traits.Firstname,
		"Lastname":       session.Identity.Traits.Lastname,
		"Nickname":       session.Identity.Traits.Nickname,
		"Birthdate":      session.Identity.Traits.Birthdate.Format(pkgVars.birthdateFormat),
		"Information":    information,
	}))
}

// Handler GET /my/profile/edit
type handleGetMyProfileEditRequestParams struct {
	cookie string
	flowID string
}

func (p *Provider) handleGetMyProfileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetMyProfileEditRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	if reqParams.flowID == "" {
		output, err := p.d.Kratos.CreateSettingsFlow(kratos.CreateSettingsFlowInput{
			Cookie: reqParams.cookie,
		})
		if err != nil {
			pkgVars.tmpl.ExecuteTemplate(w, "my/profile/edit.html", viewParameters(session, r, map[string]any{
				"ErrorMessages": output.ErrorMessages,
			}))
			return
		}
		redirect(w, r, fmt.Sprintf("%s?flow=%s", "/my/profile", output.FlowID))
		return
	}

	output, err := p.d.Kratos.GetSettingsFlow(kratos.GetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		pkgVars.tmpl.ExecuteTemplate(w, "my/profile/edit.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// セッションから現在の値を取得
	params := loadProfileFromSessionIfEmpty(updateProfileParams{}, session)

	pkgVars.tmpl.ExecuteTemplate(w, "my/profile/edit.html", viewParameters(session, r, map[string]any{
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          params.Email,
		"Firstname":      params.Firstname,
		"Lastname":       params.Lastname,
		"Nickname":       params.Nickname,
		"Birthdate":      params.Birthdate,
	}))
}

// Handler GET /my/profile/_form
type handleGetMyProfileFormRequestParams struct {
	cookie string
}

func (p *Provider) handleGetMyProfileForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetMyProfileFormRequestParams{
		cookie: r.Header.Get("Cookie"),
	}

	output, err := p.d.Kratos.CreateSettingsFlow(kratos.CreateSettingsFlowInput{
		Cookie: reqParams.cookie,
	})
	if err != nil {
		pkgVars.tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// セッションから現在の値を取得
	params := loadProfileFromSessionIfEmpty(updateProfileParams{}, session)

	pkgVars.tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          params.Email,
		"Firstname":      params.Firstname,
		"Lastname":       params.Lastname,
		"Nickname":       params.Nickname,
		"Birthdate":      params.Birthdate,
	}))
}

// Handler POST /my/profile
type handlePostMyProfileRequestPostForm struct {
	cookie    string
	flowID    string `validate:"required,uuid4"`
	csrfToken string `validate:"required"`
	Email     string `validate:"required,email" ja:"メールアドレス"`
	Firstname string `validate:"required,min=5,max=20" ja:"氏名(性)"`
	Lastname  string `validate:"required,min=5,max=20" ja:"氏名(名)"`
	Nickname  string `validate:"required,min=5,max=20" ja:"ニックネーム"`
	Birthdate string `validate:"required,birthdate" ja:"生年月日"`
}

func (p *handlePostMyProfileRequestPostForm) validate() map[string]string {
	fieldErrors := validationFieldErrors(pkgVars.validate.Struct(p))
	return fieldErrors
}

// Handler POST /my/profile
func (p *Provider) handlePostMyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handlePostMyProfileRequestPostForm{
		cookie:    r.Header.Get("Cookie"),
		flowID:    r.URL.Query().Get("flow"),
		csrfToken: r.PostFormValue("csrf_token"),
		Email:     r.PostFormValue("email"),
		Lastname:  r.PostFormValue("lastname"),
		Firstname: r.PostFormValue("firstname"),
		Nickname:  r.PostFormValue("nickname"),
		Birthdate: r.PostFormValue("birthdate"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		pkgVars.tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Email":                reqParams.Email,
			"Firstname":            reqParams.Firstname,
			"Lastname":             reqParams.Lastname,
			"Nickname":             reqParams.Nickname,
			"Birthdate":            reqParams.Birthdate,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	birthdate, err := time.Parse(pkgVars.birthdateFormat, reqParams.Birthdate)
	if err != nil {
		slog.Error(err.Error())
	}
	params := loadProfileFromSessionIfEmpty(updateProfileParams{
		FlowID:    reqParams.flowID,
		Email:     reqParams.Email,
		Firstname: reqParams.Firstname,
		Lastname:  reqParams.Lastname,
		Nickname:  reqParams.Nickname,
		Birthdate: birthdate,
	}, session)

	deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if session.NeedLoginWhenPrivilegedAccess() {
		err := saveAfterLoginHook(w, afterLoginHook{
			Operation: AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE,
			Params:    params,
		}, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
		if err != nil {
			pkgVars.tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
				"SettingsFlowID": reqParams.flowID,
				"CsrfToken":      reqParams.csrfToken,
				"ErrorMessages":  []string{"Error"},
				"Email":          params.Email,
				"Firstname":      params.Firstname,
				"Lastname":       params.Lastname,
				"Nickname":       params.Nickname,
				"Birthdate":      params.Birthdate,
			}))
		} else {
			returnTo := url.QueryEscape("/my/profile")
			slog.Info(returnTo)
			redirect(w, r, fmt.Sprintf("/auth/login?return_to=%s", returnTo))
		}
		return
	}

	// Settings Flow の送信(完了)
	output, err := p.d.Kratos.UpdateSettingsFlow(kratos.UpdateSettingsFlowInput{
		Cookie:    reqParams.cookie,
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Traits: kratos.Traits{
			Email:     params.Email,
			Firstname: params.Firstname,
			Lastname:  params.Lastname,
			Nickname:  params.Nickname,
			Birthdate: params.Birthdate,
		},
	})
	if err != nil {
		slog.Error(err.Error())
		pkgVars.tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
			"CsrfToken":     reqParams.csrfToken,
			"ErrorMessages": output.ErrorMessages,
			"Email":         params.Email,
			"Firstname":     params.Firstname,
			"Lastname":      params.Lastname,
			"Nickname":      params.Nickname,
			"Birthdate":     params.Birthdate,
		}))
		return
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	redirect(w, r, "/")
	w.WriteHeader(http.StatusOK)
}

type updateProfileParams struct {
	FlowID    string    `json:"flow_id"`
	Email     string    `json:"email"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Nickname  string    `json:"nickname"`
	Birthdate time.Time `json:"birthdate"`
}

func loadProfileFromSessionIfEmpty(params updateProfileParams, session *kratos.Session) updateProfileParams {
	if session != nil {
		if params.Email == "" {
			params.Email = session.Identity.Traits.Email
		}
		if params.Firstname == "" {
			params.Firstname = session.Identity.Traits.Firstname
		}
		if params.Lastname == "" {
			params.Lastname = session.Identity.Traits.Lastname
		}
		if params.Nickname == "" {
			params.Nickname = session.Identity.Traits.Nickname
		}
		if params.Birthdate.IsZero() {
			params.Birthdate = session.Identity.Traits.Birthdate
		}
	}
	return params
}

func (p *Provider) updateProfile(w http.ResponseWriter, r *http.Request, params updateProfileParams) error {
	ctx := r.Context()
	session := getSession(ctx)

	params = loadProfileFromSessionIfEmpty(updateProfileParams{
		Email:     params.Email,
		Firstname: params.Firstname,
		Lastname:  params.Lastname,
		Nickname:  params.Nickname,
		Birthdate: params.Birthdate,
	}, session)

	output, err := p.d.Kratos.GetSettingsFlow(kratos.GetSettingsFlowInput{
		Cookie: r.Header.Get("Cookie"),
		FlowID: params.FlowID,
	})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	// Settings Flow の送信(完了)
	updateOutput, err := p.d.Kratos.UpdateSettingsFlow(kratos.UpdateSettingsFlowInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    output.FlowID,
		CsrfToken: output.CsrfToken,
		Traits: kratos.Traits{
			Email:     params.Email,
			Firstname: params.Firstname,
			Lastname:  params.Lastname,
			Nickname:  params.Nickname,
			Birthdate: params.Birthdate,
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
