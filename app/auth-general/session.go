package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	kratosclientgo "github.com/ory/kratos-client-go"
)

func GetSession(c *gin.Context) *kratosclientgo.Session {
	session, exists := c.Get("session")
	if !exists || session == nil {
		return nil
	}
	slog.Info(fmt.Sprintf("%v", session))
	return session.(*kratosclientgo.Session)
}

func IsAuthenticated(c *gin.Context) bool {
	session := GetSession(c)
	if session != nil {
		return true
	} else {
		return false
	}
}

// セッションがprivileged_session_max_age を過ぎているかどうかを返却する
func NeedLoginWhenPrivilegedAccess(c *gin.Context) bool {
	session := GetSession(c)
	authenticateAt := session.AuthenticatedAt.In(locationJst)
	if authenticateAt.Before(time.Now().Add(-time.Minute * privilegedAccessLimitMinutes)) {
		return true
	} else {
		return false
	}
}
