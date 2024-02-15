package kratos

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	kratosclientgo "github.com/ory/kratos-client-go"
)

func readHttpResponseBody(r *http.Response) (interface{}, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return []byte{}, err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
		return []byte{}, err
	}
	return body, nil
}

// flow の ui から csrf_token を取得
func getCsrfTokenFromFlowUi(ui kratosclientgo.UiContainer) string {
	for _, node := range ui.Nodes {
		if node.Attributes.UiNodeInputAttributes.Name == "csrf_token" {
			return node.Attributes.UiNodeInputAttributes.Value.(string)
		}
	}
	return ""
}

// SuccessfulRegistration から verification flow id を取得
func getContinueWithVerificationFlowId(sr kratosclientgo.SuccessfulNativeRegistration) string {
	for _, continueWith := range sr.ContinueWith {
		if continueWith.ContinueWithVerificationUi.Action == "show_verification_ui" {
			return continueWith.ContinueWithVerificationUi.Flow.Id
		}
	}
	return ""
}

// refirect browser to を取得
func getRedirectBrowserToFromError(err kratosclientgo.ErrorBrowserLocationChangeRequired) string {
	if err.RedirectBrowserTo == nil {
		return ""
	} else {
		return *err.RedirectBrowserTo
	}
}

// kratos のエラーレスポンスからエラーメッセージを取得
func getErrorMessages(err error) []string {
	slog.Info(fmt.Sprintf("%v", err))
	oerr, ok := err.(*kratosclientgo.GenericOpenAPIError)
	if !ok {
		return []string{}
	}

	slog.Info(fmt.Sprintf("%v", oerr))
	slog.Info(fmt.Sprintf("%v", oerr.Model()))

	var messages []string
	fmt.Println(reflect.TypeOf(oerr.Model()))

	if m, ok := oerr.Model().(kratosclientgo.RegistrationFlow); ok {
		slog.Info("RegistrationFlow")
		messages = getErrorMessagesFromResigtrationFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.VerificationFlow); ok {
		slog.Info("VerificationFlow")
		messages = getErrorMessagesFromVerificationFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.LoginFlow); ok {
		slog.Info("LoginFlow")
		messages = getErrorMessagesFromLoginFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.RecoveryFlow); ok {
		slog.Info("RecoveryFlow")
		messages = getErrorMessagesFromRecoveryFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.SettingsFlow); ok {
		slog.Info("SettingsFlow")
		messages = getErrorMessagesFromSettingsFlow(m)
	}

	if m, ok := oerr.Model().(kratosclientgo.ErrorBrowserLocationChangeRequired); ok {
		slog.Info("ErrorBrowserLocationChangeRequired")
		messages = getErrorMessagesFromBrowserLocationChangeRequired(m)
	}

	if m, ok := oerr.Model().(kratosclientgo.GenericError); ok {
		slog.Info("GenericError")
		messages = getErrorMessagesFromGenericError(m)
	}

	if m, ok := oerr.Model().(kratosclientgo.ErrorGeneric); ok {
		slog.Info("GenericError")
		messages = getErrorMessagesFromErrorGeneric(m)
	}

	return messages
}

func getErrorMessagesFromResigtrationFlow(flow kratosclientgo.RegistrationFlow) []string {
	return getErrorMessagesFromUi(flow.Ui)
}

func getErrorMessagesFromVerificationFlow(flow kratosclientgo.VerificationFlow) []string {
	return getErrorMessagesFromUi(flow.Ui)
}

func getErrorMessagesFromLoginFlow(flow kratosclientgo.LoginFlow) []string {
	return getErrorMessagesFromUi(flow.Ui)
}

func getErrorMessagesFromRecoveryFlow(flow kratosclientgo.RecoveryFlow) []string {
	return getErrorMessagesFromUi(flow.Ui)
}

func getErrorMessagesFromSettingsFlow(flow kratosclientgo.SettingsFlow) []string {
	return getErrorMessagesFromUi(flow.Ui)
}

func getErrorMessagesFromUi(ui kratosclientgo.UiContainer) []string {
	slog.Info("getErrorMessagesFromUi")

	slog.Info(fmt.Sprintf("%v", ui))
	var messages []string
	for _, v := range ui.Messages {
		slog.Info(fmt.Sprintf("%v", v))
		if v.Type == "error" {
			slog.Info(fmt.Sprintf("%v", v.Id))
			slog.Info(fmt.Sprintf("%v", v.Text))
			// [TODO] 日本語化
			// https://www.ory.sh/docs/kratos/concepts/ui-user-interface#machine-readable-format
			if v.Id == 4000007 {
				messages = append(messages, "既に登録済みメールアドレスメールです")
			} else {
				messages = append(messages, v.Text)
			}
		}
	}

	return messages
}

func getErrorMessagesFromBrowserLocationChangeRequired(err kratosclientgo.ErrorBrowserLocationChangeRequired) []string {
	return getErrorMessagesFromGenericError(err.Error.Error)
}

func getErrorMessagesFromGenericError(err kratosclientgo.GenericError) []string {
	slog.Info("getErrorMessagesFromGenericError")
	if err.Id != nil {
		slog.Info(*err.Id)
		// [TODO] 日本語化
		// https://www.ory.sh/docs/kratos/concepts/ui-user-interface#ui-error-codes
		if *err.Id == "security_csrf_violation" {
			return []string{"恐れ入りますが、画面を更新してもう一度お試しください"}
		}
	}
	return []string{err.Message}
}

func getErrorMessagesFromErrorGeneric(err kratosclientgo.ErrorGeneric) []string {
	slog.Info("getErrorMessagesFromErrorGeneric")
	return getErrorMessagesFromGenericError(err.Error)
}
