package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setCookieToResponseHeader(c *gin.Context, cookies []string) {
	for _, cookie := range cookies {
		c.Writer.Header().Add("Set-Cookie", cookie)
	}
}

func redirect(c *gin.Context, redirectTo string) {
	slog.Debug(c.Request.Header.Get("HX-Request"))
	slog.Debug(redirectTo)
	if c.Request.Header.Get("HX-Request") == "true" {
		c.Writer.Header().Set("HX-Redirect", redirectTo)
		c.Status(http.StatusOK)
	} else {
		c.Redirect(303, redirectTo)
	}
}

func viewParameters(c *gin.Context, p gin.H) gin.H {
	params := p
	params["IsAuthenticated"] = isAuthenticated(c)
	params["Navbar"] = getNavbarviewParameters(c)
	params["Path"] = c.Request.URL.Path
	return params
}

func getNavbarviewParameters(c *gin.Context) gin.H {
	var nickname string

	session := getSession(c)
	if session != nil {
		nickname = session.GetValueFromTraits("nickname")
	}
	return gin.H{
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

func saveAfterLoginHook(c *gin.Context, loginHook afterLoginHook, cookieKey afterLoginHookCookieKey) error {
	cookieString, err := json.Marshal(loginHook)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	base64CookieString := base64.URLEncoding.EncodeToString(cookieString)
	slog.Info(base64CookieString)
	c.SetCookie(string(cookieKey), base64CookieString, 3600, "/", "localhost", false, true)
	return nil
}

func existsAfterLoginHook(c *gin.Context, cookieKey afterLoginHookCookieKey) bool {
	_, err := c.Cookie(string(cookieKey))
	if err != nil {
		// Cookieがない場合にエラーとはしない
		slog.Info(err.Error())
		return false
	} else {
		return true
	}
}

func loadAfterLoginHook(c *gin.Context, cookieKey afterLoginHookCookieKey) (afterLoginHook, error) {
	cookieString, err := c.Cookie(string(cookieKey))
	if err != nil {
		// Cookieがない場合にエラーとはしない
		slog.Info(err.Error())
		return afterLoginHook{}, nil
	}
	cookieBytes, err := base64.URLEncoding.DecodeString(cookieString)
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

func deleteAfterLoginHook(c *gin.Context, cookieKey afterLoginHookCookieKey) {
	c.SetCookie(string(cookieKey), "", -1, "/", "localhost", false, true)
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
