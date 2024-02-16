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

type ResponseType interface {
	kratosclientgo.RegistrationFlow |
		kratosclientgo.VerificationFlow |
		kratosclientgo.LoginFlow |
		kratosclientgo.RecoveryFlow |
		kratosclientgo.SettingsFlow |
		kratosclientgo.SuccessfulNativeRegistration |
		kratosclientgo.ErrorBrowserLocationChangeRequired
}

// goのv1.0.0のSDKには不具合があるらしく、恐らく各種flowをUnmarshalしてもnode.attributes配下を取得できない模様
// https://github.com/ory/sdk/issues/292
// 仕方ないので、interface{}型でjsonを直接パースし、そこから必要な値を取得する
func readHttpResponseBody(r *http.Response) (interface{}, error) {
	var result interface{}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
		return result, err
	}
	return result, nil
}

// goのv1.0.0のSDKには不具合があるらしく、恐らく各種flowをUnmarshalしてもnode.attributes配下を取得できない模様
// https://github.com/ory/sdk/issues/292
// 仕方ないので、interface{}型でjsonを直接パースし、そこから必要な値を取得する
// readHttpResponseBody で取得したjson(interface{})から csrf_token を取得
func getCsrfTokenFromResponseBody(responseBody interface{}) string {
	b, ok := responseBody.(map[string]interface{})
	if !ok {
		slog.Error("Fail type assertion responseBody to map[string]interface{}")
		return ""
	}
	ui, ok := b["ui"].(map[string]interface{})
	if !ok {
		slog.Error("Fail type assertion ui to map[string]interface{}")
		return ""
	}
	nodes, ok := ui["nodes"].([]interface{})
	if !ok {
		slog.Error("Fail type assertion nodes to []interface{}")
		return ""
	}

	for _, v := range nodes {
		node, ok := v.(map[string]interface{})
		if !ok {
			slog.Error("Fail type assertion node to map[string]interface{}")
			return ""
		}

		attrs, ok := node["attributes"].(map[string]interface{})
		if !ok {
			slog.Error("Fail type assertion attributes to map[string]interface{}")
			return ""
		}

		if attrs["name"] != nil && attrs["name"].(string) == "csrf_token" {
			return attrs["value"].(string)
		}
	}
	return ""
}

// goのv1.0.0のSDKには不具合があるらしく、恐らくSuccessfulNativeRegistrationをUnmarshalしてもcontinue_with配下を取得できない模様
// 関連 https://github.com/ory/sdk/issues/292
// 仕方ないので、interface{}型でjsonを直接パースし、そこから必要な値を取得する
// readHttpResponseBody で取得したjson(interface{})から continue_with.verification_ui.flow.id を取得
func getContinueWithVerificationFlowId(responseBody interface{}) string {
	b, ok := responseBody.(map[string]interface{})
	if !ok {
		slog.Error("Fail type assertion responseBody to map[string]interface{}")
		return ""
	}

	continueWith, ok := b["continue_with"].([]interface{})
	if !ok {
		slog.Error("Fail type assertion continue_with to []interface{}")
		return ""
	}

	for _, v := range continueWith {
		c, ok := v.(map[string]interface{})
		if !ok {
			slog.Error("Fail type assertion continue_with to map[string]interface{}")
			return ""
		}

		action, ok := c["action"].(string)
		if !ok {
			slog.Error("Fail type assertion action to string")
			return ""
		}

		flow, ok := c["flow"].(map[string]interface{})
		if !ok {
			slog.Error("Fail type assertion flow to map[string]interface{}")
			return ""
		}

		flowID, ok := flow["id"].(string)
		if !ok {
			slog.Error("Fail type assertion flow.id to string")
			return ""
		}

		if action == "show_verification_ui" {
			return flowID
		}
	}

	slog.Error("Missing verification flow id")
	return ""
}

// refirect browser to を取得
func getRedirectBrowserToFromError(err error) string {
	slog.Info(fmt.Sprintf("%v", err))
	oerr, ok := err.(*kratosclientgo.GenericOpenAPIError)
	if !ok {
		return ""
	}

	slog.Info(fmt.Sprintf("%v", oerr))
	slog.Info(fmt.Sprintf("%v", oerr.Model()))
	if m, ok := oerr.Model().(kratosclientgo.ErrorBrowserLocationChangeRequired); ok {
		if m.RedirectBrowserTo != nil {
			return *m.RedirectBrowserTo
		}
	}
	slog.Error("Missing redirect_browser_to")
	return ""
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
