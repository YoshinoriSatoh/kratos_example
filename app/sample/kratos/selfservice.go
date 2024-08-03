package kratos

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	PATH_SESSIONS_WHOAMI                       = "/sessions/whoami"
	PATH_SELF_SERVICE_CREATE_REGISTRATION_FLOW = "/self-service/registration/browser"
	PATH_SELF_SERVICE_UPDATE_REGISTRATION_FLOW = "/self-service/registration"
	PATH_SELF_SERVICE_GET_REGISTRATION_FLOW    = "/self-service/registration/flows"
	PATH_SELF_SERVICE_CREATE_VERIFICATION_FLOW = "/self-service/verification/browser"
	PATH_SELF_SERVICE_UPDATE_VERIFICATION_FLOW = "/self-service/verification"
	PATH_SELF_SERVICE_GET_VERIFICATION_FLOW    = "/self-service/verification/flows"
	PATH_SELF_SERVICE_CREATE_LOGIN_FLOW        = "/self-service/login/browser"
	PATH_SELF_SERVICE_UPDATE_LOGIN_FLOW        = "/self-service/login"
	PATH_SELF_SERVICE_GET_LOGIN_FLOW           = "/self-service/login/flows"
	PATH_SELF_SERVICE_GET_LOGOUT_FLOW          = "/self-service/logout/browser"
	PATH_SELF_SERVICE_UPDATE_LOGOUT_FLOW       = "/self-service/logout"
	PATH_SELF_SERVICE_CREATE_SETTINGS_FLOW     = "/self-service/settings/browser"
	PATH_SELF_SERVICE_UPDATE_SETTINGS_FLOW     = "/self-service/settings"
	PATH_SELF_SERVICE_GET_SETTINGS_FLOW        = "/self-service/settings/flows"
	PATH_SELF_SERVICE_CREATE_RECOVERY_FLOW     = "/self-service/recovery/browser"
	PATH_SELF_SERVICE_UPDATE_RECOVERY_FLOW     = "/self-service/recovery"
	PATH_SELF_SERVICE_GET_RECOVERY_FLOW        = "/self-service/recovery/flows"
	PATH_SELF_SERVICE_CALLBACK_OIDC            = "/self-service/methods/oidc/callback"
	PATH_ADMIN_LIST_IDENTITIES                 = "/admin/identities"
)

// ------------------------- Session -------------------------
type WhoamiInput struct {
	Cookie     string
	RemoteAddr string
}

type WhoamiOutput struct {
	Cookies       []string
	Session       *Session
	ErrorMessages []string
}

func (p *Provider) Whoami(i WhoamiInput) (WhoamiOutput, error) {
	var output WhoamiOutput

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       PATH_SESSIONS_WHOAMI,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	slog.Info(string(kratosOutput.BodyBytes))
	slog.Info(fmt.Sprintf("%v", kratosOutput))

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	var session Session
	if err := json.Unmarshal(kratosOutput.BodyBytes, &session); err != nil {
		slog.Error(err.Error())
		return output, err
	}
	output.Session = &session

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	slog.Info(fmt.Sprintf("%v", output.Session))
	slog.Info(fmt.Sprintf("%v", output.Session.Identity))

	return output, nil
}

// ------------------------- Registration Flow -------------------------
type RegistrationRenderingType string

const (
	RegistrationRenderingTypePassword = RegistrationRenderingType("password")
	RegistrationRenderingTypeOidc     = RegistrationRenderingType("oidc")
)

type GetRegistrationFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type GetRegistrationFlowOutput struct {
	Cookies           []string
	FlowID            string
	RenderingType     RegistrationRenderingType
	Traits            Traits
	RequestFromOidc   bool
	PasskeyCreateData string
	CsrfToken         string
	ErrorMessages     []string
}

func (p *Provider) GetRegistrationFlow(i GetRegistrationFlowInput) (GetRegistrationFlowOutput, error) {
	var (
		err    error
		output GetRegistrationFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?id=%s", PATH_SELF_SERVICE_GET_REGISTRATION_FLOW, i.FlowID),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBody kratosGetRegisrationFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)
	output.RequestFromOidc = false

	// OIDC callbackの場合は、Registration flow の UIに、OIDC Providerから取得したユーザ情報をTraitsにセット
	slog.Info(kratosRespBody.RequestUrl)
	if strings.Contains(kratosRespBody.RequestUrl, PATH_SELF_SERVICE_CALLBACK_OIDC) {
		output.RequestFromOidc = true
		output.RenderingType = RegistrationRenderingTypeOidc
		var traits Traits
		for _, node := range kratosRespBody.Ui.Nodes {
			slog.Info(fmt.Sprintf("%v", node))
			if node.Attributes.Name == "traits.email" {
				traits.Email, _ = node.Attributes.Value.(string)
			}
			if node.Attributes.Name == "traits.firstname" {
				traits.Firstname, _ = node.Attributes.Value.(string)
			}
			if node.Attributes.Name == "traits.lastname" {
				traits.Lastname, _ = node.Attributes.Value.(string)
			}
			if node.Attributes.Name == "traits.nickname" {
				traits.Nickname, _ = node.Attributes.Value.(string)
			}
			if node.Attributes.Name == "traits.birthdate" {
				if birthdate, ok := node.Attributes.Value.(string); ok {
					traits.Birthdate, _ = time.Parse(pkgVars.birthdateFormat, birthdate)
				}
			}
		}
		output.Traits = traits
	} else {
		output.RenderingType = RegistrationRenderingTypePassword
		for _, node := range kratosRespBody.Ui.Nodes {
			if node.Attributes.Name == "passkey_create_data" {
				output.PasskeyCreateData = node.Attributes.Value.(string)
			}
		}
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	slog.Info(fmt.Sprintf("%v", output))

	return output, nil
}

type CreateRegistrationFlowInput struct {
	Cookie     string
	RemoteAddr string
	ReturnTo   string
}

type CreateRegistrationFlowOutput struct {
	Cookies           []string
	FlowID            string
	RenderingType     RegistrationRenderingType
	Traits            Traits
	PasskeyCreateData string
	CsrfToken         string
	ErrorMessages     []string
}

func (p *Provider) CreateRegistrationFlow(i CreateRegistrationFlowInput) (CreateRegistrationFlowOutput, error) {
	var (
		err    error
		output CreateRegistrationFlowOutput
	)

	path := PATH_SELF_SERVICE_CREATE_REGISTRATION_FLOW
	if i.ReturnTo != "" {
		path = fmt.Sprintf("%s?return_to=%s", path, i.ReturnTo)
	}
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       path,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosCreateRegisrationFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)
	output.RenderingType = RegistrationRenderingTypePassword
	for _, node := range kratosRespBody.Ui.Nodes {
		if node.Attributes.Name == "passkey_create_data" {
			output.PasskeyCreateData = node.Attributes.Value.(string)
		}
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

// Registration Flow の送信(完了)
type UpdateRegistrationFlowInput struct {
	Cookie          string
	RemoteAddr      string
	FlowID          string
	Password        string
	CsrfToken       string
	Method          string
	Provider        string
	Traits          Traits
	PasskeyRegister string
}

type UpdateRegistrationFlowOutput struct {
	Cookies            []string
	VerificationFlowID string
	RedirectBrowserTo  string
	ErrorMessages      []string
}

func (p *Provider) UpdateRegistrationFlow(i UpdateRegistrationFlowInput) (UpdateRegistrationFlowOutput, error) {
	var (
		output           UpdateRegistrationFlowOutput
		kratosInputBytes []byte
		err              error
	)

	// Update Registration Flow
	// https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/updateRegistrationFlow
	// supported method: password, oidc
	if i.Method == "password" {
		kratosInput := kratosUpdateRegistrationFlowPasswordMethodRequest{
			CsrfToken: i.CsrfToken,
			Method:    i.Method,
			Traits:    i.Traits,
			Password:  i.Password,
		}
		kratosInputBytes, err = json.Marshal(kratosInput)
		if err != nil {
			slog.Error("MarshalError", "Error", err)
			return output, err
		}
	} else if i.Method == "oidc" {
		kratosInput := kratosUpdateRegistrationFlowOidcMethodRequest{
			CsrfToken: i.CsrfToken,
			Method:    i.Method,
			Provider:  i.Provider,
			Traits:    i.Traits,
		}
		kratosInputBytes, err = json.Marshal(kratosInput)
		if err != nil {
			slog.Error("MarshalError", "Error", err)
			return output, err
		}
	} else if i.Method == "passkey" {
		kratosInput := kratosUpdateRegistrationFlowPasskeyMethodRequest{
			CsrfToken:       i.CsrfToken,
			Method:          i.Method,
			Traits:          i.Traits,
			PasskeyRegister: i.PasskeyRegister,
		}
		slog.Info("passkey input", "method", i.Method)
		kratosInputBytes, err = json.Marshal(kratosInput)
		if err != nil {
			slog.Error("MarshalError", "Error", err)
			return output, err
		}
	} else {
		slog.Error("Invalid method", "Method", i.Method)
		return output, fmt.Errorf("invalid method: %s", i.Method)
	}

	slog.Info(string(kratosInputBytes))
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s?flow=%s", PATH_SELF_SERVICE_UPDATE_REGISTRATION_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	slog.Info(fmt.Sprintf("%d", kratosOutput.StatusCode))

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateRegistrationFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else if kratosOutput.StatusCode == http.StatusUnprocessableEntity {
			var browserLocationChangeRequired errorBrowserLocationChangeRequired
			if err := json.Unmarshal(kratosOutput.BodyBytes, &browserLocationChangeRequired); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", browserLocationChangeRequired))

			// browser location changeが返却された場合は、リダイレクト先URLを設定
			output.RedirectBrowserTo = browserLocationChangeRequired.RedirectBrowserTo

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]

		return output, err
	}

	var flowPasswordResponse kratosUpdateRegisrationFlowPasswordRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &flowPasswordResponse); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	slog.Info(fmt.Sprintf("%v", flowPasswordResponse))
	for _, c := range flowPasswordResponse.ContinueWith {
		slog.Info(fmt.Sprintf("%v", c))
		if c.Action == "show_verification_ui" {
			output.VerificationFlowID = c.Flow.ID
		}
	}

	slog.Info(output.VerificationFlowID)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Verification Flow -------------------------
type GetVerificationFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type GetVerificationFlowOutput struct {
	Cookies       []string
	FlowID        string
	IsUsedFlow    bool
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) GetVerificationFlow(i GetVerificationFlowInput) (GetVerificationFlowOutput, error) {
	var (
		err    error
		output GetVerificationFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?id=%s", PATH_SELF_SERVICE_GET_VERIFICATION_FLOW, i.FlowID),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBody kratosGetVerificationFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// flow　が使用済みかチェック
	if kratosRespBody.State == "passed_challenge" {
		output.IsUsedFlow = true
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type CreateVerificationFlowInput struct {
	Cookie     string
	RemoteAddr string
	ReturnTo   string
}

type CreateVerificationFlowOutput struct {
	Cookies       []string
	FlowID        string
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) CreateVerificationFlow(i CreateVerificationFlowInput) (CreateVerificationFlowOutput, error) {
	var (
		err    error
		output CreateVerificationFlowOutput
	)

	path := PATH_SELF_SERVICE_CREATE_VERIFICATION_FLOW
	if i.ReturnTo != "" {
		path = fmt.Sprintf("%s?return_to=%s", path, i.ReturnTo)
	}
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       path,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosCreateVerificationFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

// Verification Flow の送信(完了)
type UpdateVerificationFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	Code       string
	Email      string
	CsrfToken  string
}

type UpdateVerificationFlowOutput struct {
	Cookies       []string
	ErrorMessages []string
}

func (p *Provider) UpdateVerificationFlow(i UpdateVerificationFlowInput) (UpdateVerificationFlowOutput, error) {
	var (
		output      UpdateVerificationFlowOutput
		kratosInput kratosUpdateVerificationFlowRequest
	)

	// email設定時は、Verification Flowを更新して、アカウント検証メールを送信
	// code設定時は、Verification Flowを完了
	if i.Email != "" && i.Code == "" {
		kratosInput = kratosUpdateVerificationFlowRequest{
			Method:    "code",
			Email:     i.Email,
			CsrfToken: i.CsrfToken,
		}
	} else if i.Email == "" && i.Code != "" {
		kratosInput = kratosUpdateVerificationFlowRequest{
			Method:    "code",
			Code:      i.Code,
			CsrfToken: i.CsrfToken,
		}
	} else {
		err := fmt.Errorf("parameter convination error. email: %s, code: %s", i.Email, i.Code)
		slog.Error("Parameter convination error.", "email", i.Email, "code", i.Code)
		return output, err
	}
	kratosInputBytes, err := json.Marshal(kratosInput)
	if err != nil {
		slog.Error("MarshalError", "Error", err)
		return output, err
	}

	// Verification Flow の送信(完了)
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s?flow=%s", PATH_SELF_SERVICE_UPDATE_VERIFICATION_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateVerificationFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]
	return output, nil
}

// ------------------------- Login Flow -------------------------
type GetLoginFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type GetLoginFlowOutput struct {
	Cookies             []string
	FlowID              string
	PasskeyChallenge    string
	CsrfToken           string
	ErrorMessages       []string
	DuplicateIdentifier string
}

func (p *Provider) GetLoginFlow(i GetLoginFlowInput) (GetLoginFlowOutput, error) {
	var (
		err    error
		output GetLoginFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?id=%s", PATH_SELF_SERVICE_GET_LOGIN_FLOW, i.FlowID),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBody kratosGetLoginFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	slog.Info(string(kratosOutput.BodyBytes))

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateLoginFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.DuplicateIdentifier = getDuplicateIdentifierFromUi(kratosRespBody.Ui)
	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)
	for _, node := range kratosRespBody.Ui.Nodes {
		if node.Attributes.Name == "passkey_challenge" {
			output.PasskeyChallenge = node.Attributes.Value.(string)
		}
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type CreateLoginFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	Refresh    bool
	ReturnTo   string
}

type CreateLoginFlowOutput struct {
	Cookies          []string
	FlowID           string
	PasskeyChallenge string
	CsrfToken        string
	ErrorMessages    []string
}

func (p *Provider) CreateLoginFlow(i CreateLoginFlowInput) (CreateLoginFlowOutput, error) {
	var (
		err    error
		output CreateLoginFlowOutput
	)

	path := PATH_SELF_SERVICE_CREATE_LOGIN_FLOW
	if i.ReturnTo != "" {
		path = fmt.Sprintf("%s?return_to=%s", path, i.ReturnTo)
	}
	if i.Refresh {
		path = fmt.Sprintf("%s?refresh=true", path)
	}
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       path,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosCreateLoginFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)
	for _, node := range kratosRespBody.Ui.Nodes {
		if node.Attributes.Name == "passkey_challenge" {
			output.PasskeyChallenge = node.Attributes.Value.(string)
		}
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type UpdateLoginFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	CsrfToken  string
	Identifier string
	Password   string
}

type UpdateLoginFlowOutput struct {
	Cookies           []string
	RedirectBrowserTo string
	ErrorMessages     []string
}

// Login Flow の送信(完了)
func (p *Provider) UpdateLoginFlow(i UpdateLoginFlowInput) (UpdateLoginFlowOutput, error) {
	var (
		output           UpdateLoginFlowOutput
		kratosInputBytes []byte
		err              error
	)

	kratosInput := kratosUpdateLoginFlowPasswordRequest{
		Method:     "password",
		Identifier: i.Identifier,
		Password:   i.Password,
		CsrfToken:  i.CsrfToken,
	}
	kratosInputBytes, err = json.Marshal(kratosInput)
	if err != nil {
		slog.Error("MarshalError", "Error", err)
		return output, err
	}

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s?flow=%s", PATH_SELF_SERVICE_UPDATE_LOGIN_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateLoginFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else if kratosOutput.StatusCode == http.StatusUnprocessableEntity {
			var browserLocationChangeRequired errorBrowserLocationChangeRequired
			if err := json.Unmarshal(kratosOutput.BodyBytes, &browserLocationChangeRequired); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", browserLocationChangeRequired))

			// browser location changeが返却された場合は、リダイレクト先URLを設定
			output.RedirectBrowserTo = browserLocationChangeRequired.RedirectBrowserTo

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]

		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type UpdateOidcLoginFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	CsrfToken  string
	Provider   string
}

type UpdateOidcLoginFlowOutput struct {
	Cookies           []string
	RedirectBrowserTo string
	ErrorMessages     []string
}

func (p *Provider) UpdateOidcLoginFlow(i UpdateOidcLoginFlowInput) (UpdateOidcLoginFlowOutput, error) {
	var (
		output           UpdateOidcLoginFlowOutput
		kratosInputBytes []byte
		err              error
	)

	kratosInput := kratosUpdateLoginFlowOidcRequest{
		Method:    "oidc",
		CsrfToken: i.CsrfToken,
		Provider:  i.Provider,
	}
	kratosInputBytes, err = json.Marshal(kratosInput)
	if err != nil {
		slog.Error("MarshalError", "Error", err)
		return output, err
	}

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s?flow=%s", PATH_SELF_SERVICE_UPDATE_LOGIN_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateLoginFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else if kratosOutput.StatusCode == http.StatusUnprocessableEntity {
			var browserLocationChangeRequired errorBrowserLocationChangeRequired
			if err := json.Unmarshal(kratosOutput.BodyBytes, &browserLocationChangeRequired); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", browserLocationChangeRequired))

			// browser location changeが返却された場合は、リダイレクト先URLを設定
			output.RedirectBrowserTo = browserLocationChangeRequired.RedirectBrowserTo

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]

		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Logout Flow -------------------------
type LogoutFlowInput struct {
	Cookie     string
	RemoteAddr string
}

type LogoutFlowOutput struct {
	Cookies       []string
	ErrorMessages []string
}

func (p *Provider) Logout(i LogoutFlowInput) (LogoutFlowOutput, error) {
	var (
		output LogoutFlowOutput
		err    error
	)

	// create flow
	kratosOutputCreateFlow, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       PATH_SELF_SERVICE_GET_LOGOUT_FLOW,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBodyCreateFlow kratosCreateLogoutFlowRespnse
	if err := json.Unmarshal(kratosOutputCreateFlow.BodyBytes, &kratosRespBodyCreateFlow); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutputCreateFlow.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutputCreateFlow.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutputCreateFlow.Header["Set-Cookie"]
		return output, err
	}

	// update flow
	kratosOutputUpdateFlow, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?flow=%s&token=%s", PATH_SELF_SERVICE_UPDATE_LOGOUT_FLOW, kratosRespBodyCreateFlow.ID, kratosRespBodyCreateFlow.LogoutToken),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBodyUpdateFlow kratosUpdateLogoutFlowRequest
	if err := json.Unmarshal(kratosOutputUpdateFlow.BodyBytes, &kratosRespBodyUpdateFlow); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutputUpdateFlow.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutputUpdateFlow.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutputUpdateFlow.Header["Set-Cookie"]
		return output, err
	}

	output.Cookies = kratosOutputUpdateFlow.Header["Set-Cookie"]
	return output, nil
}

// ------------------------- Recovery Flow -------------------------
type GetRecoveryFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type GetRecoveryFlowOutput struct {
	Cookies       []string
	FlowID        string
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) GetRecoveryFlow(i GetRecoveryFlowInput) (GetRecoveryFlowOutput, error) {
	var (
		err    error
		output GetRecoveryFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?id=%s", PATH_SELF_SERVICE_GET_RECOVERY_FLOW, i.FlowID),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosGetRecoveryFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type CreateRecoveryFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type CreateRecoveryFlowOutput struct {
	Cookies       []string
	FlowID        string
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) CreateRecoveryFlow(i CreateRecoveryFlowInput) (CreateRecoveryFlowOutput, error) {
	var (
		err    error
		output CreateRecoveryFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       PATH_SELF_SERVICE_CREATE_RECOVERY_FLOW,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosCreateRecoveryFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type UpdateRecoveryFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	CsrfToken  string
	Email      string
	Code       string
}

type UpdateRecoveryFlowOutput struct {
	Cookies           []string
	RedirectBrowserTo string
	ErrorMessages     []string
}

// Recovery Flow の送信(完了)
func (p *Provider) UpdateRecoveryFlow(i UpdateRecoveryFlowInput) (UpdateRecoveryFlowOutput, error) {
	var (
		output      UpdateRecoveryFlowOutput
		kratosInput kratosUpdateRecoveryFlowRequest
	)

	// email設定時は、Recovery Flowを更新して、アカウント復旧メールを送信
	// code設定時は、Recovery Flowを完了
	if i.Email != "" && i.Code == "" {
		kratosInput = kratosUpdateRecoveryFlowRequest{
			Method:    "code",
			Email:     i.Email,
			CsrfToken: i.CsrfToken,
		}
	} else if i.Email == "" && i.Code != "" {
		kratosInput = kratosUpdateRecoveryFlowRequest{
			Method:    "code",
			Code:      i.Code,
			CsrfToken: i.CsrfToken,
		}
	} else {
		err := fmt.Errorf("parameter convination error. email: %s, code: %s", i.Email, i.Code)
		slog.Error("Parameter convination error.", "email", i.Email, "code", i.Code)
		return output, err
	}
	kratosInputBytes, err := json.Marshal(kratosInput)
	if err != nil {
		slog.Error("MarshalError", "Error", err)
		return output, err
	}

	// Verification Flow の送信(完了)
	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s?flow=%s", PATH_SELF_SERVICE_GET_RECOVERY_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateSettingsFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else if kratosOutput.StatusCode == http.StatusUnprocessableEntity {
			var browserLocationChangeRequired errorBrowserLocationChangeRequired
			if err := json.Unmarshal(kratosOutput.BodyBytes, &browserLocationChangeRequired); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", browserLocationChangeRequired))

			// browser location changeが返却された場合は、リダイレクト先URLを設定
			output.RedirectBrowserTo = browserLocationChangeRequired.RedirectBrowserTo

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]

		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

// ------------------------- Settings Flow -------------------------
type GetSettingsFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type GetSettingsFlowOutput struct {
	Cookies       []string
	FlowID        string
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) GetSettingsFlow(i GetSettingsFlowInput) (GetSettingsFlowOutput, error) {
	var (
		err    error
		output GetSettingsFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("%s?id=%s", PATH_SELF_SERVICE_GET_SETTINGS_FLOW, i.FlowID),
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}
	var kratosRespBody kratosGetSettingsFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type CreateSettingsFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
}

type CreateSettingsFlowOutput struct {
	Cookies       []string
	FlowID        string
	CsrfToken     string
	ErrorMessages []string
}

func (p *Provider) CreateSettingsFlow(i CreateSettingsFlowInput) (CreateSettingsFlowOutput, error) {
	var (
		err    error
		output CreateSettingsFlowOutput
	)

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodGet,
		Path:       PATH_SELF_SERVICE_CREATE_SETTINGS_FLOW,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	var kratosRespBody kratosCreateSettingsFlowRespnse
	if err := json.Unmarshal(kratosOutput.BodyBytes, &kratosRespBody); err != nil {
		slog.Error(err.Error())
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		output.Cookies = kratosOutput.Header["Set-Cookie"]
		return output, err
	}

	output.FlowID = kratosRespBody.ID
	output.CsrfToken = getCsrfTokenFromFlowUi(kratosRespBody.Ui)

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type UpdateSettingsFlowInput struct {
	Cookie     string
	RemoteAddr string
	FlowID     string
	CsrfToken  string
	Method     string
	Password   string
	Traits     Traits
}

type UpdateSettingsFlowOutput struct {
	Cookies           []string
	RedirectBrowserTo string
	ErrorMessages     []string
}

// Settings Flow (password) の送信(完了)
func (p *Provider) UpdateSettingsFlow(i UpdateSettingsFlowInput) (UpdateSettingsFlowOutput, error) {
	var (
		output      UpdateSettingsFlowOutput
		kratosInput kratosUpdateSettingsFlowRequest
		err         error
	)

	if i.Method == "password" {
		kratosInput = kratosUpdateSettingsFlowRequest{
			CsrfToken: i.CsrfToken,
			Method:    i.Method,
			Password:  i.Password,
		}
	} else if i.Method == "profile" {
		kratosInput = kratosUpdateSettingsFlowRequest{
			CsrfToken: i.CsrfToken,
			Method:    i.Method,
			Traits:    i.Traits,
		}
	} else {
		err := fmt.Errorf("invalid method: %s", i.Method)
		slog.Error(err.Error())
		return output, err
	}

	kratosInputBytes, err := json.Marshal(kratosInput)
	if err != nil {
		slog.Error("MarshalError", "Error", err)
		return output, err
	}

	kratosOutput, err := p.requestKratosPublic(requestKratosInput{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("%s[?flow=%s", PATH_SELF_SERVICE_UPDATE_SETTINGS_FLOW, i.FlowID),
		BodyBytes:  kratosInputBytes,
		Cookie:     i.Cookie,
		RemoteAddr: i.RemoteAddr,
	})
	if err != nil {
		slog.Error("requestKratosPublic error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		if kratosOutput.StatusCode == http.StatusBadRequest {
			// status code 400 の場合のレスポンスボディのフォーマットは複数存在する
			var flow kratosUpdateSettingsFlowBadRequestErrorResponse
			if err := json.Unmarshal(kratosOutput.BodyBytes, &flow); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", flow))
			if flow.Error != nil {
				output.ErrorMessages = getErrorMessagesFromGenericError(*flow.Error)
			} else if flow.Ui != nil {
				output.ErrorMessages = getErrorMessagesFromUi(*flow.Ui)
			} else {
				slog.Info("Unknown error response format")
			}

		} else if kratosOutput.StatusCode == http.StatusUnprocessableEntity {
			var browserLocationChangeRequired errorBrowserLocationChangeRequired
			if err := json.Unmarshal(kratosOutput.BodyBytes, &browserLocationChangeRequired); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", browserLocationChangeRequired))

			// browser location changeが返却された場合は、リダイレクト先URLを設定
			output.RedirectBrowserTo = browserLocationChangeRequired.RedirectBrowserTo

		} else {
			var errGeneric errorGeneric
			if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
				slog.Error(err.Error())
				return output, err
			}
			slog.Info(fmt.Sprintf("%v", errGeneric))
			output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		}
		output.Cookies = kratosOutput.Header["Set-Cookie"]

		return output, err
	}

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type AdminGetIdentityInput struct {
	ID                string `json:"id"`
	IncludeCredential string `json:"include_credential"`
	Cookie            string `json:"cookie"`
}

type AdminGetIdentityOutput struct {
	Cookies       []string
	Identity      Identity `json:"identity"`
	ErrorMessages []string `json:"error_messages"`
}

func (p *Provider) AdminGetIdentity(i AdminGetIdentityInput) (AdminGetIdentityOutput, error) {
	var (
		output AdminGetIdentityOutput
		err    error
	)

	kratosOutput, err := p.requestKratosAdmin(requestKratosInput{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/admin/identities/%s?include_credential=%s", i.ID, i.IncludeCredential),
		// Cookie: i.Cookie,
	})
	if err != nil {
		slog.Error("requestKratosAdmin error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
	}

	var identity Identity
	if err := json.Unmarshal(kratosOutput.BodyBytes, &identity); err != nil {
		slog.Error(err.Error())
		return output, err
	}
	output.Identity = identity

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}

type AdminListIdentitiesInput struct {
	Cookie               string `json:"cookie"`
	CredentialIdentifier string `json:"credential_identifier"`
}

type AdminListIdentitiesOutput struct {
	Cookies       []string
	Identities    []Identity `json:"identities"`
	ErrorMessages []string   `json:"error_messages"`
}

func (p *Provider) AdminListIdentities(i AdminListIdentitiesInput) (AdminListIdentitiesOutput, error) {
	var (
		output AdminListIdentitiesOutput
		err    error
	)

	slog.Debug("AdminListIdentities", "input", i)

	kratosOutput, err := p.requestKratosAdmin(requestKratosInput{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("%s?credential_identifier=%s", PATH_ADMIN_LIST_IDENTITIES, i.CredentialIdentifier),
		// Cookie: i.Cookie,
	})
	if err != nil {
		slog.Error("requestKratosAdmin error", "Error", err)
		return output, err
	}

	// error handling
	if kratosOutput.StatusCode != http.StatusOK {
		slog.Debug("AdminListIdentities failed", "kratosOutput", kratosOutput)
		slog.Debug(string(kratosOutput.BodyBytes))
		var errGeneric errorGeneric
		if err := json.Unmarshal(kratosOutput.BodyBytes, &errGeneric); err != nil {
			slog.Error(err.Error())
			return output, err
		}
		slog.Info(fmt.Sprintf("%v", errGeneric))
		output.ErrorMessages = getErrorMessagesFromGenericError(errGeneric.Error)
		return output, err
	}

	slog.Debug("AdminListIdentities succeeded", "kratosOutput", kratosOutput)
	slog.Debug(string(kratosOutput.BodyBytes))

	var identities []Identity
	if err := json.Unmarshal(kratosOutput.BodyBytes, &identities); err != nil {
		slog.Error(err.Error())
		return output, err
	}
	output.Identities = identities

	// browser flowでは、kartosから受け取ったcookieをそのままブラウザへ返却する
	output.Cookies = kratosOutput.Header["Set-Cookie"]

	return output, nil
}
