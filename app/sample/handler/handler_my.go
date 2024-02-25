package handler

import (
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
	"net/url"
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

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})

	if err != nil {
		tmpl.ExecuteTemplate(w, "my/password/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	tmpl.ExecuteTemplate(w, "my/password/index.html", viewParameters(session, r, map[string]any{
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
	fieldErrors := validationFieldErrors(validate.Struct(p))
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
		tmpl.ExecuteTemplate(w, "my/password/_form.html", viewParameters(session, r, map[string]any{
			"SettingsFlowID":       reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Password":             reqParams.password,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	// Setting Flow 更新
	output, err := p.d.Kratos.UpdateSettingsFlowPassword(kratos.UpdateSettingsFlowPasswordInput{
		Cookie:    r.Header.Get("Cookie"),
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Password:  reqParams.password,
	})
	if err != nil {
		slog.Info(err.Error())
		tmpl.ExecuteTemplate(w, "my/password/_form.html", viewParameters(session, r, map[string]any{
			"SettingsFlowID": reqParams.flowID,
			"CsrfToken":      reqParams.csrfToken,
			"Password":       reqParams.password,
			"ErrorMessages":  output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	redirect(w, r, "/auth/login")
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

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "my/profile/index.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// flowの情報に従ってレンダリング
	email := session.GetValueFromTraits("email")
	firstname := session.GetValueFromTraits("firstname")
	lastname := session.GetValueFromTraits("lastname")
	nickname := session.GetValueFromTraits("nickname")
	birthdate := session.GetValueFromTraits("birthdate")
	var information string
	if existsAfterLoginHook(r, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE) {
		information = "プロフィールを更新しました。"
		deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
	}
	tmpl.ExecuteTemplate(w, "my/profile/index.html", viewParameters(session, r, map[string]any{
		"SettingsFlowID": output.FlowID,
		"CsrfToken":      output.CsrfToken,
		"Email":          email,
		"Firstname":      firstname,
		"Lastname":       lastname,
		"Nickname":       nickname,
		"Birthdate":      birthdate,
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

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "my/profile/edit.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// セッションから現在の値を取得
	params := loadProfileFromSessionIfEmpty(updateProfileParams{}, session)

	tmpl.ExecuteTemplate(w, "my/profile/edit.html", viewParameters(session, r, map[string]any{
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
	flowID string
}

func (p *Provider) handleGetMyProfileForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	reqParams := handleGetMyProfileFormRequestParams{
		cookie: r.Header.Get("Cookie"),
		flowID: r.URL.Query().Get("flow"),
	}

	// Setting Flow の作成 or 取得
	// Setting flowを新規作成した場合は、FlowIDを含めてリダイレクト
	output, err := p.d.Kratos.CreateOrGetSettingsFlow(kratos.CreateOrGetSettingsFlowInput{
		Cookie: reqParams.cookie,
		FlowID: reqParams.flowID,
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
			"ErrorMessages": output.ErrorMessages,
		}))
	}

	// kratosのcookieをそのままブラウザへ受け渡す
	setCookieToResponseHeader(w, output.Cookies)

	// セッションから現在の値を取得
	params := loadProfileFromSessionIfEmpty(updateProfileParams{}, session)

	tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
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
	Birthdate string `validate:"required,datetime=2006-01-02" ja:"生年月日"`
}

func (p *handlePostMyProfileRequestPostForm) validate() map[string]string {
	fieldErrors := validationFieldErrors(validate.Struct(p))
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
		Nickname:  r.PostFormValue("nickname"),
		Birthdate: r.PostFormValue("birthdate"),
	}
	validationFieldErrors := reqParams.validate()
	if len(validationFieldErrors) > 0 {
		tmpl.ExecuteTemplate(w, "auth/registration/_form.html", viewParameters(session, r, map[string]any{
			"RegistrationFlowID":   reqParams.flowID,
			"CsrfToken":            reqParams.csrfToken,
			"Firstname":            reqParams.Firstname,
			"Lastname":             reqParams.Lastname,
			"Nickname":             reqParams.Nickname,
			"Birthdate":            reqParams.Birthdate,
			"ValidationFieldError": validationFieldErrors,
		}))
		return
	}

	params := loadProfileFromSessionIfEmpty(updateProfileParams{
		FlowID:    reqParams.flowID,
		Email:     reqParams.Email,
		Firstname: reqParams.Firstname,
		Lastname:  reqParams.Lastname,
		Nickname:  reqParams.Nickname,
		Birthdate: reqParams.Birthdate,
	}, session)

	deleteAfterLoginHook(w, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)

	// セッションが privileged_session_max_age を過ぎていた場合、ログイン画面へリダイレクト（再ログインの強制）
	if session.NeedLoginWhenPrivilegedAccess() {
		err := saveAfterLoginHook(w, afterLoginHook{
			Operation: AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE,
			Params:    params,
		}, AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE)
		if err != nil {
			tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
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
			redirect(w, r, "/auth/login")
		}
		return
	}

	// Settings Flow の送信(完了)
	output, err := p.d.Kratos.UpdateSettingsFlowProfile(kratos.UpdateSettingsFlowProfileInput{
		Cookie:    reqParams.cookie,
		FlowID:    reqParams.flowID,
		CsrfToken: reqParams.csrfToken,
		Traits: map[string]interface{}{
			"email":     params.Email,
			"firstname": params.Firstname,
			"lastname":  params.Lastname,
			"nickname":  params.Nickname,
			"birthdate": params.Birthdate,
		},
	})
	if err != nil {
		tmpl.ExecuteTemplate(w, "my/profile/_form.html", viewParameters(session, r, map[string]any{
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
	FlowID    string `json:"flow_id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Nickname  string `json:"nickname"`
	Birthdate string `json:"birthdate"`
}

func loadProfileFromSessionIfEmpty(params updateProfileParams, session *kratos.Session) updateProfileParams {
	if session != nil {
		if params.Email == "" {
			params.Email = session.GetValueFromTraits("email")
		}
		if params.Firstname == "" {
			params.Firstname = session.GetValueFromTraits("firstname")
		}
		if params.Lastname == "" {
			params.Lastname = session.GetValueFromTraits("lastname")
		}
		if params.Nickname == "" {
			params.Nickname = session.GetValueFromTraits("nickname")
		}
		if params.Birthdate == "" {
			params.Birthdate = session.GetValueFromTraits("birthdate")
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
			"firstname": params.Firstname,
			"lastname":  params.Lastname,
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
