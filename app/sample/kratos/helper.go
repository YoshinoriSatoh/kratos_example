package kratos

import (
	"fmt"
	"log/slog"
	"time"
)

// // goのv1.0.0のSDKには不具合があるらしく、恐らく各種flowをUnmarshalしてもnode.attributes配下を取得できない模様
// // https://github.com/ory/sdk/issues/292
// // 仕方ないので、interface{}型でjsonを直接パースし、そこから必要な値を取得する
// func readHttpResponseBody(r *http.Response) (interface{}, error) {
// 	var result interface{}

// 	defer r.Body.Close()
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return result, err
// 	}

// 	if err := json.Unmarshal(body, &result); err != nil {
// 		slog.Error(err.Error())
// 		return result, err
// 	}
// 	return result, nil
// }

// func getVerificationFlowIDFromSuccessfulNativeRegistration(s *kratosclientgo.SuccessfulNativeRegistration) string {
// 	for _, c := range s.ContinueWith {
// 		return c.ContinueWithVerificationUi.Flow.Id
// 	}
// 	slog.Error("Missing csrf_token")
// 	return ""
// }

func getCsrfTokenFromFlowUi(ui uiContainer) string {
	for _, node := range ui.Nodes {
		if node.Attributes.Name == "csrf_token" {
			return node.Attributes.Value.(string)
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

// // readHttpResponseBody で取得したjson(interface{})から continue_with.verification_ui.flow.id を取得
// func getContinueWithVerificationFlowId(responseBody interface{}) string {
// 	b, ok := responseBody.(map[string]interface{})
// 	if !ok {
// 		slog.Error("Fail type assertion responseBody to map[string]interface{}")
// 		return ""
// 	}

// 	continueWith, ok := b["continue_with"].([]interface{})
// 	if !ok {
// 		slog.Error("Fail type assertion continue_with to []interface{}")
// 		return ""
// 	}

// 	for _, v := range continueWith {
// 		c, ok := v.(map[string]interface{})
// 		if !ok {
// 			slog.Error("Fail type assertion continue_with to map[string]interface{}")
// 			return ""
// 		}

// 		action, ok := c["action"].(string)
// 		if !ok {
// 			slog.Error("Fail type assertion action to string")
// 			return ""
// 		}

// 		flow, ok := c["flow"].(map[string]interface{})
// 		if !ok {
// 			slog.Error("Fail type assertion flow to map[string]interface{}")
// 			return ""
// 		}

// 		flowID, ok := flow["id"].(string)
// 		if !ok {
// 			slog.Error("Fail type assertion flow.id to string")
// 			return ""
// 		}

// 		if action == "show_verification_ui" {
// 			return flowID
// 		}
// 	}

// 	slog.Error("Missing verification flow id")
// 	return ""
// }

// // refirect browser to を取得 (SDKバグ回避暫定用)
// func getRedirectBrowserToFromHttpResponse(r *http.Response) string {
// 	var e errorBrowserLocationChangeRequired

// 	defer r.Body.Close()
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return ""
// 	}

// 	if err := json.Unmarshal(body, &e); err != nil {
// 		slog.Error(err.Error())
// 		return ""
// 	}

// 	return e.RedirectBrowserTo
// }

// // refirect browser to を取得 (本来はこちらを使用したい)
// func getRedirectBrowserToFromError(err error) string {
// 	slog.Info(fmt.Sprintf("%v", err))
// 	oerr, ok := err.(*kratosclientgo.GenericOpenAPIError)
// 	if !ok {
// 		return ""
// 	}

// 	slog.Info(fmt.Sprintf("%v", oerr))
// 	slog.Info(fmt.Sprintf("%v", oerr.Model()))
// 	if m, ok := oerr.Model().(kratosclientgo.ErrorBrowserLocationChangeRequired); ok {
// 		if m.RedirectBrowserTo != nil {
// 			return *m.RedirectBrowserTo
// 		}
// 	}
// 	slog.Error("Missing redirect_browser_to")
// 	return ""
// }

// // kratos のエラーレスポンスからエラーメッセージを取得
// func getErrorMessages(err error) []string {
// 	slog.Info(fmt.Sprintf("%v", err))
// 	oerr, ok := err.(*kratosclientgo.GenericOpenAPIError)
// 	if !ok {
// 		return []string{}
// 	}

// 	slog.Info(fmt.Sprintf("%v", oerr))
// 	slog.Info(fmt.Sprintf("%v", oerr.Model()))

// 	var messages []string
// 	fmt.Println(reflect.TypeOf(oerr.Model()))

// 	if m, ok := oerr.Model().(kratosUpdateRegistrationFlowBadRequestErrorResponse); ok {
// 		slog.Info("RegistrationFlow")
// 		messages = getErrorMessagesFromResigtrationFlow(m)
// 	}
// 	if m, ok := oerr.Model().(verificationFlow); ok {
// 		slog.Info("VerificationFlow")
// 		messages = getErrorMessagesFromVerificationFlow(m)
// 	}
// 	if m, ok := oerr.Model().(loginFlow); ok {
// 		slog.Info("LoginFlow")
// 		messages = getErrorMessagesFromLoginFlow(m)
// 	}
// 	if m, ok := oerr.Model().(recoveryFlow); ok {
// 		slog.Info("RecoveryFlow")
// 		messages = getErrorMessagesFromRecoveryFlow(m)
// 	}
// 	if m, ok := oerr.Model().(settingsFlow); ok {
// 		slog.Info("SettingsFlow")
// 		messages = getErrorMessagesFromSettingsFlow(m)
// 	}

// 	// if m, ok := oerr.Model().(errorBrowserLocationChangeRequired); ok {
// 	// 	slog.Info("ErrorBrowserLocationChangeRequired")
// 	// 	messages = getErrorMessagesFromBrowserLocationChangeRequired(m)
// 	// }

// 	if m, ok := oerr.Model().(genericError); ok {
// 		slog.Info("GenericError")
// 		messages = getErrorMessagesFromGenericError(m)
// 	}

// 	if m, ok := oerr.Model().(errorGeneric); ok {
// 		slog.Info("GenericError")
// 		messages = getErrorMessagesFromErrorGeneric(m)
// 	}

// 	return messages
// }

// func getErrorMessagesFromResigtrationFlow(flow kratosUpdateRegistrationFlowBadRequestErrorResponse) []string {
// 	return getErrorMessagesFromUi(*flow.Ui)
// }

// func getErrorMessagesFromVerificationFlow(flow verificationFlow) []string {
// 	return getErrorMessagesFromUi(flow.Ui)
// }

// func getErrorMessagesFromLoginFlow(flow loginFlow) []string {
// 	return getErrorMessagesFromUi(flow.Ui)
// }

// func getErrorMessagesFromRecoveryFlow(flow recoveryFlow) []string {
// 	return getErrorMessagesFromUi(flow.Ui)
// }

// func getErrorMessagesFromSettingsFlow(flow settingsFlow) []string {
// 	return getErrorMessagesFromUi(flow.Ui)
// }

func getErrorMessagesFromUi(ui uiContainer) []string {
	slog.Info("getErrorMessagesFromUi")

	slog.Info(fmt.Sprintf("%v", ui))
	var messages []string
	for _, v := range ui.Messages {
		slog.Info(fmt.Sprintf("%v", v))
		if v.Type == "error" {
			slog.Info(fmt.Sprintf("%v", v.ID))
			slog.Info(fmt.Sprintf("%v", v.Text))
			// [TODO] 日本語化
			// https://www.ory.sh/docs/kratos/concepts/ui-user-interface#machine-readable-format
			if v.ID == 4000007 {
				messages = append(messages, "既に登録済みメールアドレスメールです")
			} else {
				messages = append(messages, v.Text)
			}
		}
	}

	return messages
}

func getDuplicateIdentifierFromUi(ui uiContainer) string {
	slog.Info(fmt.Sprintf("%v", ui))
	for _, v := range ui.Messages {
		slog.Info(fmt.Sprintf("%v", v))
		if v.ID == 1010016 && v.Type == "info" {
			slog.Info(fmt.Sprintf("%v", v.ID))
			slog.Info(fmt.Sprintf("%v", v.Text))
			return v.Context["duplicateIdentifier"].(string)
		}
	}

	return ""
}

// func getErrorMessagesFromBrowserLocationChangeRequired(err errorBrowserLocationChangeRequired) []string {
// 	return getErrorMessagesFromGenericError(err.Error)
// }

func getErrorMessagesFromGenericError(err genericError) []string {
	slog.Info("getErrorMessagesFromGenericError")
	slog.Info(err.ID)
	// [TODO] 日本語化
	// https://www.ory.sh/docs/kratos/concepts/ui-user-interface#ui-error-codes
	if err.ID == "security_csrf_violation" {
		return []string{"恐れ入りますが、画面を更新してもう一度お試しください"}
	}
	return []string{err.Message}
}

func getErrorMessagesFromErrorGeneric(err errorGeneric) []string {
	slog.Info("getErrorMessagesFromErrorGeneric")
	return getErrorMessagesFromGenericError(err.Error)
}

// セッションがprivileged_session_max_age を過ぎているかどうかを返却する
func (s *Session) NeedLoginWhenPrivilegedAccess() bool {
	authenticateAt := s.AuthenticatedAt.In(pkgVars.locationJst)
	if authenticateAt.Before(time.Now().Add(-time.Second * pkgVars.privilegedAccessLimitMinutes)) {
		return true
	} else {
		return false
	}
}
