package kratos

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
		return "", err
	}
	return result.(map[string]interface{})["redirect_browser_to"].(string), nil
}

func getErrorMessagesFromFlowHttpResponse(r *http.Response) []string {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		return []string{}
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error(err.Error())
		return []string{}
	}

	var messages []string
	resultObj, ok := result.(map[string]interface{})
	slog.Info(fmt.Sprintf("%v", resultObj))
	if ok && resultObj != nil {
		uiObj, uiOk := resultObj["ui"].(map[string]interface{})
		slog.Info(fmt.Sprintf("%v", uiObj))
		errorObj, errorOk := resultObj["error"].(map[string]interface{})
		slog.Info(fmt.Sprintf("%v", errorObj))
		if uiOk && uiObj != nil {
			messageArr, ok := uiObj["messages"].([]interface{})
			slog.Info(fmt.Sprintf("%v", messageArr...))
			if ok && len(messageArr) > 0 {
				for _, message := range messageArr {
					slog.Info(fmt.Sprintf("%v", message))
					if message.(map[string]interface{})["type"].(string) == "error" {
						slog.Info(fmt.Sprintf("%v", message))
						messages = append(messages, message.(map[string]interface{})["text"].(string))
					}
				}
			}
		} else if errorOk && errorObj != nil {
			messages = append(messages, errorObj["message"].(string))
		}
	}
	return messages
}

// type LogoutIfNeededInput struct {
// 	Cookie  string
// 	Session *kratosclientgo.Session
// }

// func LogoutIfNeeded(i LogoutIfNeededInput) (bool, error) {
// 	for _, v := range i.Session.AuthenticationMethods {
// 		if *v.Method == "code_recovery" {
// 			_, err := Logout(LogoutFlowInput{
// 				Cookie: i.Cookie,
// 			})
// 			if err != nil {
// 				slog.Error(err.Error())
// 				return false, err
// 			} else {
// 				return true, nil
// 			}
// 		}
// 	}
// 	return false, nil
// }
