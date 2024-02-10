package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	kratosclientgo "github.com/ory/kratos-client-go"
)

func ViewParameters(c *gin.Context, p gin.H) gin.H {
	merged := p
	merged["IsAuthenticated"] = IsAuthenticated(c)
	merged["Navbar"] = getNavbarViewParameters(c)
	return merged
}

func getNavbarViewParameters(c *gin.Context) gin.H {
	var nickname string

	session := GetSession(c)
	if session != nil {
		nickname, _ = session.Identity.Traits.(map[string]interface{})["nickname"].(string)
	}
	return gin.H{
		"Nickname": nickname,
	}
}

// ------------------------- Settings profile edit view paremeter -------------------------
type settingsProfileEditViewParams struct {
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Birthdate string `json:"birthdate"`
}

const SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY = "settings_profile_edit_view_params"

func mergeSettingsProfileEditViewParams(params settingsProfileEditViewParams, session *kratosclientgo.Session) settingsProfileEditViewParams {
	var ok bool
	if session != nil {
		if params.Email == "" {
			params.Email, ok = session.Identity.Traits.(map[string]interface{})["email"].(string)
			if !ok {
				params.Email = ""
			}
		}
		if params.Nickname == "" {
			params.Nickname, ok = session.Identity.Traits.(map[string]interface{})["nickname"].(string)
			if !ok {
				params.Nickname = ""
			}
		}
		if params.Birthdate == "" {
			params.Birthdate, ok = session.Identity.Traits.(map[string]interface{})["birthdate"].(string)
			if !ok {
				params.Birthdate = ""
			}
		}
	}
	slog.Info(fmt.Sprintf("%v", params))
	return params
}

// Cookieからプロフィール編集画面のパラメータを取得
func loadSettingsProfileEditParamsFromCookie(c *gin.Context) (settingsProfileEditViewParams, error) {
	editSettingParamStr, err := c.Cookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY)
	if err != nil {
		slog.Error(err.Error())
		return settingsProfileEditViewParams{}, err
	}
	editSettingParams, err := base64.URLEncoding.DecodeString(editSettingParamStr)
	if err != nil {
		slog.Error(err.Error())
		return settingsProfileEditViewParams{}, err
	}

	var params settingsProfileEditViewParams
	err = json.Unmarshal(editSettingParams, &params)
	if err != nil {
		slog.Error(err.Error())
		return settingsProfileEditViewParams{}, err
	}

	c.SetCookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY, "", -1, "/", "localhost", false, true)

	return params, nil
}

// プロフィール編集画面のパラメータをCookieに保存
func saveSettingsProfileEditParamsToCookie(c *gin.Context, params settingsProfileEditViewParams) error {
	cookieString, err := json.Marshal(params)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	base64CookieString := base64.URLEncoding.EncodeToString(cookieString)
	c.SetCookie(SETTINGS_PROFILE_EDIT_VIEW_PARAMS_COOKIE_KEY, base64CookieString, 3600, "/", "localhost", false, true)
	return nil
}
