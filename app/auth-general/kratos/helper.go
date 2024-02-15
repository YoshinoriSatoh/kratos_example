package kratos

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	kratosclientgo "github.com/ory/kratos-client-go"
)

// flow の ui から csrf_token を取得
// SDKを使用しているので、本来は上記レスポンスの第一引数である registrationFlow *kratosclientgo.RegistrationFlow から取得するところだが、
// goのv1.0.0のSDKには不具合があるらしく、仕方ないのでhttp.Responseから取得している
// https://github.com/ory/sdk/issues/292
func getCsrfTokenFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var csrfToken string
	for _, node := range result.(map[string]interface{})["ui"].(map[string]interface{})["nodes"].([]interface{}) {
		attrName := node.(map[string]interface{})["attributes"].(map[string]interface{})["name"]
		if attrName != nil && attrName.(string) == "csrf_token" {
			csrfToken = node.(map[string]interface{})["attributes"].(map[string]interface{})["value"].(string)
			break
		}
	}
	slog.Info(csrfToken)
	return csrfToken, nil
}

func getContinueWithVerificationUiFlowIdFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var verificationFlowID string
	for _, continueWith := range result.(map[string]interface{})["continue_with"].([]interface{}) {
		if continueWith.(map[string]interface{})["action"].(string) == "show_verification_ui" {
			verificationFlowID = continueWith.(map[string]interface{})["flow"].(map[string]interface{})["id"].(string)
			break
		}
	}
	slog.Info(verificationFlowID)
	return verificationFlowID, nil
}

func getRedirectBrowserToFromFlowHttpResponse(r *http.Response) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
	}
	return result.(map[string]interface{})["redirect_browser_to"].(string), nil
}

func getErrorMessages(err error) []string {
	slog.Info(fmt.Sprintf("%v", err))
	oerr, ok := err.(*kratosclientgo.GenericOpenAPIError)
	if !ok {
		return []string{}
	}

	slog.Info(fmt.Sprintf("%v", oerr))
	slog.Info(fmt.Sprintf("%v", oerr.Model()))

	var messages []string

	if m, ok := oerr.Model().(kratosclientgo.RegistrationFlow); ok {
		messages = getErrorMessagesFromResigtrationFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.VerificationFlow); ok {
		messages = getErrorMessagesFromVerificationFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.LoginFlow); ok {
		messages = getErrorMessagesFromLoginFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.RecoveryFlow); ok {
		messages = getErrorMessagesFromRecoveryFlow(m)
	}
	if m, ok := oerr.Model().(kratosclientgo.SettingsFlow); ok {
		messages = getErrorMessagesFromSettingsFlow(m)
	}

	if m, ok := oerr.Model().(kratosclientgo.ErrorBrowserLocationChangeRequired); ok {
		messages = getErrorMessagesFromBrowserLocationChangeRequired(m)
	}

	if m, ok := oerr.Model().(kratosclientgo.GenericError); ok {
		messages = getErrorMessagesFromGenericError(m)
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
			// [TODO] ここは日本語化しないといけない
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
	slog.Info("%v", err)
	if err.Id != nil {
		slog.Info(*err.Id)
	}
	return []string{err.Message}
}
