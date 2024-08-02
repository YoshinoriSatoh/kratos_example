package kratos

import "time"

type Traits struct {
	Email     string    `json:"email" validate:"required,email" ja:"メールアドレス"`
	Firstname string    `json:"firstname" validate:"required" ja:"氏名(姓)"`
	Lastname  string    `json:"lastname" validate:"required" ja:"氏名(名)"`
	Nickname  string    `json:"nickname" validate:"required" ja:"ニックネーム"`
	Birthdate time.Time `json:"birthdate" ja:"生年月日"`
}

type Identity struct {
	ID     string `json:"id" validate:"required"`
	Traits Traits `json:"traits" validate:"required"`
}

// session
type Session struct {
	ID              string    `json:"id"`
	Identity        Identity  `json:"identity,omitempty"`
	AuthenticatedAt time.Time `json:"authenticated_at"`
}

// kratosからのレスポンスのうち、必要なもののみを定義

type uiText struct {
	Context map[string]interface{} `json:"context,omitempty"`
	ID      int64                  `json:"id"`
	Text    string                 `json:"text"`
	Type    string                 `json:"type"`
}

type uiNodeMeta struct {
	Label uiText `json:"label,omitempty"`
}

type uiNodeAttributes struct {
	Disabled bool        `json:"disabled"`
	Label    uiText      `json:"label,omitempty"`
	Name     string      `json:"name"`
	NodeType string      `json:"node_type"`
	Onclick  string      `json:"onclick,omitempty"`
	Pattern  string      `json:"pattern,omitempty"`
	Required bool        `json:"required,omitempty"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value,omitempty"`
}

type uiNode struct {
	Attributes uiNodeAttributes `json:"attributes"`
	Group      string           `json:"group"`
	Messages   []uiText         `json:"messages"`
	Meta       uiNodeMeta       `json:"meta"`
	Type       string           `json:"type"`
}

type uiContainer struct {
	Action   string   `json:"action"`
	Messages []uiText `json:"messages,omitempty"`
	Method   string   `json:"method"`
	Nodes    []uiNode `json:"nodes"`
}

type verificationFlow struct {
	Ui uiContainer `json:"ui"`
}

type loginFlow struct {
	Ui uiContainer `json:"ui"`
}

type recoveryFlow struct {
	Ui uiContainer `json:"ui"`
}

type settingsFlow struct {
	Ui uiContainer `json:"ui"`
}

type genericError struct {
	Code    int64                  `json:"code,omitempty"`
	Debug   string                 `json:"debug,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	ID      string                 `json:"id,omitempty"`
	Message string                 `json:"message"`
	Reason  string                 `json:"reason,omitempty"`
	Request string                 `json:"request,omitempty"`
	Status  string                 `json:"status,omitempty"`
}

type errorGeneric struct {
	Error genericError `json:"error"`
}

type errorBrowserLocationChangeRequired struct {
	RedirectBrowserTo string `json:"redirect_browser_to,omitempty"`
}

type continueWithFlow struct {
	ID string `json:"id"`
}

type continueWith struct {
	Action string           `json:"action"`
	Flow   continueWithFlow `json:"flow"`
}

// Registration flow
type kratosCreateRegisrationFlowRespnse struct {
	ID         string      `json:"id"`
	Ui         uiContainer `json:"ui"`
	RequestUrl string      `json:"request_url"`
}

type kratosGetRegisrationFlowRespnse struct {
	ID         string      `json:"id"`
	Ui         uiContainer `json:"ui"`
	RequestUrl string      `json:"request_url"`
}

type kratosUpdateRegistrationFlowPasswordMethodRequest struct {
	CsrfToken string `json:"csrf_token"`
	Method    string `json:"method"`
	Traits    Traits `json:"traits"`
	Password  string `json:"password"`
}

type kratosUpdateRegistrationFlowOidcMethodRequest struct {
	CsrfToken string `json:"csrf_token"`
	Method    string `json:"method"`
	Provider  string `json:"provider"`
	Traits    Traits `json:"traits"`
}

type kratosUpdateRegistrationFlowPasskeyMethodRequest struct {
	CsrfToken       string `json:"csrf_token"`
	Method          string `json:"method"`
	Traits          Traits `json:"traits"`
	PasskeyRegister string `json:"passkey_register"`
}

type kratosUpdateRegisrationFlowPasswordRespnse struct {
	ContinueWith []continueWith `json:"continue_with"`
}

// status code 400 の場合のレスポンスボディのフォーマット
// ドキュメントではregistration flowが返却される記載しかないが、GenericErrorが返却される場合もある
// どちらの場合にも対応するため、必要なフィールドを全て定義している
type kratosUpdateRegistrationFlowBadRequestErrorResponse struct {
	Ui    *uiContainer  `json:"ui,omitempty"`
	Error *genericError `json:"error,omitempty"`
}

// Verification flow
type kratosCreateVerificationFlowRespnse struct {
	ID string      `json:"id"`
	Ui uiContainer `json:"ui"`
}

type kratosGetVerificationFlowRespnse struct {
	ID    string      `json:"id"`
	Ui    uiContainer `json:"ui"`
	State string      `json:"state"`
}

type kratosUpdateVerificationFlowRequest struct {
	Method    string `json:"method"`
	Email     string `json:"email"`
	Code      string `json:"code"`
	CsrfToken string `json:"csrf_token"`
}

// status code 400 の場合のレスポンスボディのフォーマット
// ドキュメントではverification flowが返却される記載しかないが、GenericErrorが返却される場合もある
// どちらの場合にも対応するため、必要なフィールドを全て定義している
type kratosUpdateVerificationFlowBadRequestErrorResponse struct {
	Ui    *uiContainer  `json:"ui,omitempty"`
	Error *genericError `json:"error,omitempty"`
}

// Login flow
type kratosCreateLoginFlowRespnse struct {
	ID string      `json:"id"`
	Ui uiContainer `json:"ui"`
}

type kratosGetLoginFlowRespnse struct {
	ID    string      `json:"id"`
	Ui    uiContainer `json:"ui"`
	State string      `json:"state"`
}

type kratosUpdateLoginFlowPasswordRequest struct {
	Method     string `json:"method"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	CsrfToken  string `json:"csrf_token"`
}

type kratosUpdateLoginFlowOidcRequest struct {
	Method    string `json:"method"`
	Provider  string `json:"provider"`
	CsrfToken string `json:"csrf_token"`
}

// status code 400 の場合のレスポンスボディのフォーマット
// ドキュメントではverification flowが返却される記載しかないが、GenericErrorが返却される場合もある
// どちらの場合にも対応するため、必要なフィールドを全て定義している
type kratosUpdateLoginFlowBadRequestErrorResponse struct {
	Ui    *uiContainer  `json:"ui,omitempty"`
	Error *genericError `json:"error,omitempty"`
}

// Logout flow
type kratosCreateLogoutFlowRespnse struct {
	ID          string `json:"id"`
	LogoutToken string `json:"logout_token"`
}

type kratosUpdateLogoutFlowRequest struct {
	CsrfToken string `json:"csrf_token"`
}

// Recovery flow
type kratosCreateRecoveryFlowRespnse struct {
	ID string      `json:"id"`
	Ui uiContainer `json:"ui"`
}

type kratosGetRecoveryFlowRespnse struct {
	ID string      `json:"id"`
	Ui uiContainer `json:"ui"`
}

type kratosUpdateRecoveryFlowRequest struct {
	Method    string `json:"method"`
	Email     string `json:"email"`
	Code      string `json:"code"`
	CsrfToken string `json:"csrf_token"`
}

// Settings flow
type kratosCreateSettingsFlowRespnse struct {
	ID string      `json:"id"`
	Ui uiContainer `json:"ui"`
}

type kratosGetSettingsFlowRespnse struct {
	ID    string      `json:"id"`
	Ui    uiContainer `json:"ui"`
	State string      `json:"state"`
}

type kratosUpdateSettingsFlowRequest struct {
	Method    string `json:"method"`
	Password  string `json:"password"`
	Traits    Traits `json:"traits"`
	Code      string `json:"code"`
	CsrfToken string `json:"csrf_token"`
}

// status code 400 の場合のレスポンスボディのフォーマット
// ドキュメントではverification flowが返却される記載しかないが、GenericErrorが返却される場合もある
// どちらの場合にも対応するため、必要なフィールドを全て定義している
type kratosUpdateSettingsFlowBadRequestErrorResponse struct {
	Ui    *uiContainer  `json:"ui,omitempty"`
	Error *genericError `json:"error,omitempty"`
}
