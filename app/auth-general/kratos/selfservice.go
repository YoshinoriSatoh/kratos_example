package kratos

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	kratosclientgo "github.com/ory/kratos-client-go"
)

type Identity struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Birthdate time.Time `json:"birthdate"`
}

// ------------------------- Session -------------------------
type ToSessionInput struct {
	Cookie string
}

type ToSessionOutput struct {
	Cookies       []string
	Session       *Session
	ErrorMessages []string
}

func (p *Provider) ToSession(i ToSessionInput) (ToSessionOutput, error) {
	var output ToSessionOutput
	slog.Info("ToSession", "Cookie", i.Cookie)
	session, response, err := p.kratosPublicClient.FrontendApi.
		ToSession(context.Background()).
		Cookie(i.Cookie).
		Execute()
	if err != nil {
		slog.Info("Unauthorized", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}

	output.Session = NewFromKratosClientSession(session)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Registration Flow -------------------------
type CreateOrGetRegistrationFlowInput struct {
	Cookie string
	FlowID string
}

type CreateOrGetRegistrationFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsNewFlow     bool
	CsrfToken     string
	ErrorMessages []string
}

// Registration Flow がなければ新規作成、あれば取得
// csrfTokenは、本来は *kratosclientgo.RegistrationFlow から取得できるはずだが、
// kratos-client-go:v1.0.0 に不具合があるため、http.Response から取得し返却している
func (p *Provider) CreateOrGetRegistrationFlow(i CreateOrGetRegistrationFlowInput) (CreateOrGetRegistrationFlowOutput, error) {
	var (
		err              error
		response         *http.Response
		registrationFlow *kratosclientgo.RegistrationFlow
		output           CreateOrGetRegistrationFlowOutput
	)

	// flowID がない場合は新規にRegistration Flow を作成
	// flowID がある場合はRegistration Flow を取得
	if i.FlowID == "" {
		registrationFlow, response, err = p.kratosPublicClient.FrontendApi.
			CreateBrowserRegistrationFlow(context.Background()).
			Execute()
		if err != nil {
			slog.Error("CreateRegistrationFlow Error", "RegistrationFlow", registrationFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("CreateRegistrationFlow Succeed", "RegistrationFlow", registrationFlow, "Response", response)

		output.IsNewFlow = true

	} else {
		registrationFlow, response, err = p.kratosPublicClient.FrontendApi.
			GetRegistrationFlow(context.Background()).
			Id(i.FlowID).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("GetRegistrationFlow Error", "RegistrationFlow", registrationFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("GetRegisrationFlow Succeed", "RegistrationFlow", registrationFlow, "Response", response)
	}

	output.FlowID = registrationFlow.Id

	// flow の ui から csrf_token を取得
	output.CsrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error("Get csrf_token from http response Error", "Response", response, "Error", err)
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// Registration Flow の送信(完了)
type UpdateRegistrationFlowInput struct {
	Cookie    string
	FlowID    string
	Email     string
	Password  string
	CsrfToken string
}

type UpdateRegistrationFlowOutput struct {
	Cookies            []string
	VerificationFlowID string
	ErrorMessages      []string
}

func (p *Provider) UpdateRegistrationFlow(i UpdateRegistrationFlowInput) (UpdateRegistrationFlowOutput, error) {
	var output UpdateRegistrationFlowOutput

	// Registration Flow の送信(完了)
	updateRegistrationFlowBody := kratosclientgo.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &kratosclientgo.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: i.Password,
			Traits: map[string]interface{}{
				"email": i.Email,
			},
			CsrfToken: &i.CsrfToken,
		},
	}
	successfulRegistration, response, err := p.kratosPublicClient.FrontendApi.
		UpdateRegistrationFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateRegistrationFlowBody(updateRegistrationFlowBody).
		Execute()
	if err != nil {
		slog.Error("UpdateRegistrationFlow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}
	slog.Info("UpdateRegisrationFlow Succeed", "SuccessfulRegistration", successfulRegistration, "Response", response)

	output.VerificationFlowID, err = getContinueWithVerificationUiFlowIdFromFlowHttpResponse(response)
	if err != nil {
		slog.Error("UpdateRegistrationFlow Error", "Response", response, "Error", err)
		output.ErrorMessages = []string{err.Error()}
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Verification Flow -------------------------
type CreateOrGetVerificationFlowInput struct {
	Cookie string
	FlowID string
}

type CreateOrGetVerificationFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsNewFlow     bool
	IsUsedFlow    bool
	CsrfToken     string
	ErrorMessages []string
}

// Verification Flow がなければ新規作成、あれば取得
// csrfTokenは、本来は *kratosclientgo.VerificationFlow から取得できるはずだが、
// kratos-client-go:v1.0.0 に不具合があるため、http.Response から取得し返却している
func (p *Provider) CreateOrGetVerificationFlow(i CreateOrGetVerificationFlowInput) (CreateOrGetVerificationFlowOutput, error) {
	var (
		err              error
		response         *http.Response
		verificationFlow *kratosclientgo.VerificationFlow
		output           CreateOrGetVerificationFlowOutput
	)

	// flowID がない場合は新規にVerification Flow を作成
	// flowID がある場合はVerification Flow を取得
	if i.FlowID == "" {
		verificationFlow, response, err = p.kratosPublicClient.FrontendApi.
			CreateBrowserVerificationFlow(context.Background()).
			Execute()
		if err != nil {
			slog.Error("CreateVerificationFlow Error", "VerificationFlow", verificationFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("CreateVerificationFlow Succeed", "VerificationFlow", verificationFlow, "Response", response)

		output.IsNewFlow = true

	} else {
		verificationFlow, response, err = p.kratosPublicClient.FrontendApi.
			GetVerificationFlow(context.Background()).
			Id(i.FlowID).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("Get Verification Flow Error", "VerificationFlow", verificationFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("GetVerificationFlow Succeed", "VerificationFlow", verificationFlow, "Response", response)
	}

	output.FlowID = verificationFlow.Id

	// flow　が使用済みかチェック
	if verificationFlow.State == kratosclientgo.VERIFICATIONFLOWSTATE_PASSED_CHALLENGE {
		output.IsUsedFlow = true
	}

	// flow の ui から csrf_token を取得
	output.CsrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// // Verification Flow の送信(完了)
type UpdateVerificationFlowInput struct {
	Cookie    string
	FlowID    string
	Code      string
	Email     string
	CsrfToken string
}

type UpdateVerificationFlowOutput struct {
	Cookies       []string
	ErrorMessages []string
}

func (p *Provider) UpdateVerificationFlow(i UpdateVerificationFlowInput) (UpdateVerificationFlowOutput, error) {
	var (
		output     UpdateVerificationFlowOutput
		updateBody kratosclientgo.UpdateVerificationFlowWithCodeMethod
	)

	// email設定時は、Verification Flowを更新して、アカウント検証メールを送信
	// code設定時は、Verification Flowを完了
	if i.Email != "" && i.Code == "" {
		updateBody = kratosclientgo.UpdateVerificationFlowWithCodeMethod{
			Method:    "code",
			Email:     &i.Email,
			CsrfToken: &i.CsrfToken,
		}
	} else if i.Email == "" && i.Code != "" {
		updateBody = kratosclientgo.UpdateVerificationFlowWithCodeMethod{
			Method:    "code",
			Code:      &i.Code,
			CsrfToken: &i.CsrfToken,
		}
	} else {
		err := fmt.Errorf("parameter convination error. email: %s, code: %s", i.Email, i.Code)
		slog.Error("Parameter convination error.", "email", i.Email, "code", i.Code)
		return output, err
	}

	// Verification Flow の送信(完了)
	updateVerificationFlowBody := kratosclientgo.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &updateBody,
	}
	successfulVerification, response, err := p.kratosPublicClient.FrontendApi.
		UpdateVerificationFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateVerificationFlowBody(updateVerificationFlowBody).
		Execute()
	if err != nil {
		slog.Error("UpdateVerificationFlow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, nil
	}
	slog.Info("UpdateVerification Succeed", "SuccessfulVerification", successfulVerification, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Login Flow -------------------------
type CreateOrGetLoginFlowInput struct {
	Cookie  string
	FlowID  string
	Refresh bool
}

type CreateOrGetLoginFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsNewFlow     bool
	CsrfToken     string
	ErrorMessages []string
}

// Login Flow がなければ新規作成、あれば取得
// csrfTokenは、本来は *kratosclientgo.LoginFlow から取得できるはずだが、
// kratos-client-go:v1.0.0 に不具合があるため、http.Response から取得し返却している
func (p *Provider) CreateOrGetLoginFlow(i CreateOrGetLoginFlowInput) (CreateOrGetLoginFlowOutput, error) {
	var (
		err       error
		response  *http.Response
		loginFlow *kratosclientgo.LoginFlow
		output    CreateOrGetLoginFlowOutput
	)

	// flowID がない場合は新規にLogin Flow を作成
	// flowID がある場合はLogin Flow を取得
	if i.FlowID == "" {
		loginFlow, response, err = p.kratosPublicClient.FrontendApi.
			CreateBrowserLoginFlow(context.Background()).
			Refresh(i.Refresh).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("CreateLoginFlow Error", "LoginFlow", loginFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("CreateLoginFlow Succeed", "LoginFlow", loginFlow, "Response", response)

		output.IsNewFlow = true

	} else {
		loginFlow, response, err = p.kratosPublicClient.FrontendApi.
			GetLoginFlow(context.Background()).
			Id(i.FlowID).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("GetLoginFlow Error", "LoginFlow", loginFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("GetLoginFlow Succeed", "LoginFlow", loginFlow, "Response", response)
	}

	output.FlowID = loginFlow.Id

	// flow の ui から csrf_token を取得
	output.CsrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error("Get csrf_token from http response Error", "Response", response, "Error", err)
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

type UpdateLoginFlowInput struct {
	Cookie     string
	FlowID     string
	CsrfToken  string
	Identifier string
	Password   string
}

type UpdateLoginFlowOutput struct {
	Cookies       []string
	ErrorMessages []string
}

// Login Flow の送信(完了)
func (p *Provider) UpdateLoginFlow(i UpdateLoginFlowInput) (UpdateLoginFlowOutput, error) {
	var output UpdateLoginFlowOutput

	updateLoginFlowBody := kratosclientgo.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &kratosclientgo.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: i.Identifier,
			Password:   i.Password,
			CsrfToken:  &i.CsrfToken,
		},
	}
	successfulLogin, response, err := p.kratosPublicClient.FrontendApi.
		UpdateLoginFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateLoginFlowBody(updateLoginFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Login Flow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}
	slog.Info("UpdateLoginFlow Succeed", "SuccessfulLogin", successfulLogin, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Logout Flow -------------------------
type LogoutFlowInput struct {
	Cookie string
}

type LogoutFlowOutput struct {
	Cookies       []string
	ErrorMessages []string
}

func (p *Provider) Logout(i LogoutFlowInput) (LogoutFlowOutput, error) {
	var output LogoutFlowOutput

	logoutFlow, response, err := p.kratosPublicClient.FrontendApi.
		CreateBrowserLogoutFlow(context.Background()).
		Cookie(i.Cookie).
		Execute()
	if err != nil {
		slog.Error("CreateLogoutFlow Error", "LogoutFlow", logoutFlow, "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}

	// Logout Flow の送信(完了)
	response, err = p.kratosPublicClient.FrontendApi.
		UpdateLogoutFlow(context.Background()).
		Token(logoutFlow.LogoutToken).
		Cookie(i.Cookie).
		Execute()
	if err != nil {
		slog.Error("UpdateLogout Flow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}
	slog.Info("UpdateLoginFlow Succeed", "Response", response)
	output.Cookies = response.Header["Set-Cookie"]
	return output, nil
}

// ------------------------- Recovery Flow -------------------------
type CreateOrGetRecoveryFlowInput struct {
	Cookie string
	FlowID string
}

type CreateOrGetRecoveryFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsNewFlow     bool
	CsrfToken     string
	ErrorMessages []string
}

// Recovery Flow がなければ新規作成、あれば取得
// csrfTokenは、本来は *kratosclientgo.RecoveryFlow から取得できるはずだが、
// kratos-client-go:v1.0.0 に不具合があるため、http.Response から取得し返却している
func (p *Provider) CreateOrGetRecoveryFlow(i CreateOrGetRecoveryFlowInput) (CreateOrGetRecoveryFlowOutput, error) {
	var (
		err          error
		response     *http.Response
		recoveryFlow *kratosclientgo.RecoveryFlow
		output       CreateOrGetRecoveryFlowOutput
	)

	// flowID がない場合は新規にRecovery Flow を作成してリダイレクト
	// flowID がある場合はRecovery Flow を取得
	if i.FlowID == "" {
		recoveryFlow, response, err = p.kratosPublicClient.FrontendApi.
			CreateBrowserRecoveryFlow(context.Background()).
			Execute()
		if err != nil {
			slog.Error("CreateRecoveryFlow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("CreateRecoveryFlo Succeed", "RecoveryFlow", recoveryFlow, "Response", response)

		output.IsNewFlow = true

	} else {
		recoveryFlow, response, err = p.kratosPublicClient.FrontendApi.
			GetRecoveryFlow(context.Background()).
			Id(i.FlowID).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("GetRecoveryFlow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("GetRecoveryFlow Succeed", "RecoveryFlow", recoveryFlow, "Response", response)
	}

	output.FlowID = recoveryFlow.Id

	// flow の ui から csrf_token を取得
	output.CsrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error("Get csrf_token from http response Error", "Response", response, "Error", err)
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

type UpdateRecoveryFlowInput struct {
	Cookie    string
	FlowID    string
	CsrfToken string
	Email     string
	Code      string
}

type UpdateRecoveryFlowOutput struct {
	Cookies           []string
	RedirectBrowserTo string
	ErrorMessages     []string
}

// Recovery Flow の送信(完了)
func (p *Provider) UpdateRecoveryFlow(i UpdateRecoveryFlowInput) (UpdateRecoveryFlowOutput, error) {
	var (
		output     UpdateRecoveryFlowOutput
		updateBody kratosclientgo.UpdateRecoveryFlowWithCodeMethod
	)

	// email設定時は、Recovery Flowを更新して、アカウント復旧メールを送信
	// code設定時は、Recovery Flowを完了
	if i.Email != "" && i.Code == "" {
		updateBody = kratosclientgo.UpdateRecoveryFlowWithCodeMethod{
			Method:    "code",
			Email:     &i.Email,
			CsrfToken: &i.CsrfToken,
		}
	} else if i.Email == "" && i.Code != "" {
		updateBody = kratosclientgo.UpdateRecoveryFlowWithCodeMethod{
			Method:    "code",
			Code:      &i.Code,
			CsrfToken: &i.CsrfToken,
		}
	} else {
		err := fmt.Errorf("parameter convination error. email: %s, code: %s", i.Email, i.Code)
		slog.Error("Parameter convination error.", "email", i.Email, "code", i.Code)
		return output, err
	}

	// Recovery Flow を更新
	updateRecoveryFlowBody := kratosclientgo.UpdateRecoveryFlowBody{
		UpdateRecoveryFlowWithCodeMethod: &updateBody,
	}
	recoveryFlow, response, err := p.kratosPublicClient.FrontendApi.
		UpdateRecoveryFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateRecoveryFlowBody(updateRecoveryFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Recovery Flow Error", "RecoveryFlow", recoveryFlow, "Response", response, "Error", err)
		// browser location changeが返却された場合は、リダイレクト先URLを設定
		if response.StatusCode == 422 {
			output.Cookies = response.Header["Set-Cookie"]
			output.RedirectBrowserTo, _ = getRedirectBrowserToFromFlowHttpResponse(response)
		} else {
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		}
		return output, err
	}
	slog.Info("UpdateRecovery Succeed", "RecoveryFlow", recoveryFlow, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Settings Flow -------------------------
type CreateOrGetSettingsFlowInput struct {
	Cookie string
	FlowID string
}

type CreateOrGetSettingsFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsNewFlow     bool
	CsrfToken     string
	ErrorMessages []string
}

// Settings Flow がなければ新規作成、あれば取得
// csrfTokenは、本来は *kratosclientgo.SettingsFlow から取得できるはずだが、
// kratos-client-go:v1.0.0 に不具合があるため、http.Response から取得し返却している
func (p *Provider) CreateOrGetSettingsFlow(i CreateOrGetSettingsFlowInput) (CreateOrGetSettingsFlowOutput, error) {
	var (
		err          error
		response     *http.Response
		settingsFlow *kratosclientgo.SettingsFlow
		output       CreateOrGetSettingsFlowOutput
	)

	// flowID がない場合は新規にSettings Flow を作成してリダイレクト
	// flowID がある場合はSettings Flow を取得
	if i.FlowID == "" {
		settingsFlow, response, err = p.kratosPublicClient.FrontendApi.
			CreateBrowserSettingsFlow(context.Background()).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("CreateSettingsFlow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("CreateSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)

		output.IsNewFlow = true

	} else {
		settingsFlow, response, err = p.kratosPublicClient.FrontendApi.
			GetSettingsFlow(context.Background()).
			Id(i.FlowID).
			Cookie(i.Cookie).
			Execute()
		if err != nil {
			slog.Error("GetSettingsFlow Error", "SettingsFlow", settingsFlow, "Response", response, "Error", err)
			output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
			return output, err
		}
		slog.Info("GetSettingsFlow Succeed", "SettingsFlow", settingsFlow, "Response", response)
	}

	output.FlowID = settingsFlow.Id

	// flow の ui から csrf_token を取得
	output.CsrfToken, err = getCsrfTokenFromFlowHttpResponse(response)
	if err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

type UpdateSettingsFlowPasswordInput struct {
	Cookie    string
	FlowID    string
	CsrfToken string
	Password  string
}

type UpdateSettingsFlowPasswordOutput struct {
	Cookies       []string
	ErrorMessages []string
}

// Settings Flow (password) の送信(完了)
func (p *Provider) UpdateSettingsFlowPassword(i UpdateSettingsFlowPasswordInput) (UpdateSettingsFlowPasswordOutput, error) {
	var (
		output UpdateSettingsFlowPasswordOutput
	)

	// Settings Flow の送信(完了)
	updateSettingsFlowBody := kratosclientgo.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithPasswordMethod: &kratosclientgo.UpdateSettingsFlowWithPasswordMethod{
			Method:    "password",
			Password:  i.Password,
			CsrfToken: &i.CsrfToken,
		},
	}
	successfulSettings, response, err := p.kratosPublicClient.FrontendApi.
		UpdateSettingsFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateSettingsFlowBody(updateSettingsFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Settings Flow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}
	slog.Info("UpdateSettings Succeed", "SuccessfulSettings", successfulSettings, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}

type UpdateSettingsFlowProfileInput struct {
	Cookie    string
	FlowID    string
	CsrfToken string
	Traits    map[string]interface{}
}

type UpdateSettingsFlowProfileOutput struct {
	Cookies       []string
	ErrorMessages []string
}

// Settings Flow (profile) の送信(完了)
func (p *Provider) UpdateSettingsFlowProfile(i UpdateSettingsFlowProfileInput) (UpdateSettingsFlowProfileOutput, error) {
	var (
		output UpdateSettingsFlowProfileOutput
	)

	// Settings Flow の送信(完了)
	updateSettingsFlowBody := kratosclientgo.UpdateSettingsFlowBody{
		UpdateSettingsFlowWithProfileMethod: &kratosclientgo.UpdateSettingsFlowWithProfileMethod{
			Method:    "profile",
			Traits:    i.Traits,
			CsrfToken: &i.CsrfToken,
		},
	}
	successfulSettings, response, err := p.kratosPublicClient.FrontendApi.
		UpdateSettingsFlow(context.Background()).
		Flow(i.FlowID).
		Cookie(i.Cookie).
		UpdateSettingsFlowBody(updateSettingsFlowBody).
		Execute()
	if err != nil {
		slog.Error("Update Settings Flow Error", "Response", response, "Error", err)
		output.ErrorMessages = getErrorMessagesFromFlowHttpResponse(response)
		return output, err
	}
	slog.Info("UpdateSettings Succeed", "SuccessfulSettings", successfulSettings, "Response", response)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = response.Header["Set-Cookie"]

	return output, nil
}
