package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"
)

func setCookieToResponseHeader(w http.ResponseWriter, cookies []string) {
	for _, cookie := range cookies {
		w.Header().Add("Set-Cookie", cookie)
	}
}

func redirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	slog.Debug(r.Header.Get("HX-Request"))
	slog.Debug(redirectTo)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirectTo)
		w.WriteHeader(http.StatusSeeOther)
	} else {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}

func viewParameters(session *kratos.Session, r *http.Request, p map[string]any) map[string]any {
	params := p
	params["IsAuthenticated"] = isAuthenticated(session)
	params["Navbar"] = getNavbarviewParameters(session)
	params["Path"] = r.URL.Path
	return params
}

func getNavbarviewParameters(session *kratos.Session) map[string]any {
	var nickname string

	if session != nil {
		nickname = session.GetValueFromTraits("nickname")
	}
	return map[string]any{
		"Nickname": nickname,
	}
}

// -------------------------
type afterLoginHook struct {
	Operation afterLoginHookOperation `json:"operation"`
	Params    interface{}             `json:"params"`
}

type afterLoginHookOperation string

const (
	AFTER_LOGIN_HOOK_OPERATION_UPDATE_PROFILE = "update_profile"
)

type afterLoginHookCookieKey string

const (
	AFTER_LOGIN_HOOK_COOKIE_KEY_SETTINGS_PROFILE_UPDATE = "after_login_hook_settings_profile_update"
)

func saveAfterLoginHook(w http.ResponseWriter, loginHook afterLoginHook, cookieKey afterLoginHookCookieKey) error {
	cookieString, err := json.Marshal(loginHook)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	base64CookieString := base64.URLEncoding.EncodeToString(cookieString)
	slog.Info(base64CookieString)
	http.SetCookie(w, &http.Cookie{
		Name:     string(cookieKey),
		Value:    base64CookieString,
		MaxAge:   3600,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
	})
	return nil
}

func existsAfterLoginHook(r *http.Request, cookieKey afterLoginHookCookieKey) bool {
	_, err := r.Cookie(string(cookieKey))
	if err != nil {
		// Cookieがない場合にエラーとはしない
		slog.Info(err.Error())
		return false
	} else {
		return true
	}
}

func loadAfterLoginHook(r *http.Request, cookieKey afterLoginHookCookieKey) (afterLoginHook, error) {
	cookieString, err := r.Cookie(string(cookieKey))
	if err != nil {
		// Cookieがない場合にエラーとはしない
		slog.Info(err.Error())
		return afterLoginHook{}, nil
	}
	cookieBytes, err := base64.URLEncoding.DecodeString(cookieString.Value)
	if err != nil {
		slog.Error(err.Error())
		return afterLoginHook{}, err
	}

	var hook afterLoginHook
	err = json.Unmarshal(cookieBytes, &hook)
	if err != nil {
		slog.Error(err.Error())
		return afterLoginHook{}, err
	}

	return hook, nil
}

func deleteAfterLoginHook(w http.ResponseWriter, cookieKey afterLoginHookCookieKey) {
	http.SetCookie(w, &http.Cookie{
		Name:     string(cookieKey),
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
	})
}

// ------------------------- Settings profile edit view paremeter -------------------------
type settingsProfileEditViewParams struct {
	FlowID    string `json:"flow_id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Birthdate string `json:"birthdate"`
}

const SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY = "settings_profile_edit_view_params"

func mergeSettingsProfileEditViewParams(params settingsProfileEditViewParams, session *kratos.Session) settingsProfileEditViewParams {
	if session != nil {
		if params.Email == "" {
			params.Email = session.GetValueFromTraits("email")
		}
		if params.Nickname == "" {
			params.Nickname = session.GetValueFromTraits("nickname")
		}
		if params.Birthdate == "" {
			params.Birthdate = session.GetValueFromTraits("birthdate")
		}
	}
	slog.Info(fmt.Sprintf("%v", params))
	return params
}

// // Cookieからプロフィール編集画面のパラメータを取得
// func loadSettingsProfileEditParamsFromCookie(c *gin.Context) (settingsProfileEditViewParams, error) {
// 	editSettingParamStr, err := c.Cookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY)
// 	if err != nil {
// 		// Cookieがない場合にエラーとはしない
// 		slog.Info(err.Error())
// 		return settingsProfileEditViewParams{}, nil
// 	}
// 	editSettingParams, err := base64.URLEncoding.DecodeString(editSettingParamStr)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return settingsProfileEditViewParams{}, err
// 	}

// 	var params settingsProfileEditViewParams
// 	err = json.Unmarshal(editSettingParams, &params)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return settingsProfileEditViewParams{}, err
// 	}

// 	c.SetCookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY, "", -1, "/", "localhost", false, true)

// 	return params, nil
// }

// // プロフィール編集画面のパラメータをCookieに保存
// func saveSettingsProfileEditParamsToCookie(c *gin.Context, params settingsProfileEditViewParams) error {
// 	cookieString, err := json.Marshal(params)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return err
// 	}
// 	base64CookieString := base64.URLEncoding.EncodeToString(cookieString)
// 	c.SetCookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY, base64CookieString, 3600, "/", "localhost", false, true)
// 	return nil
// }
