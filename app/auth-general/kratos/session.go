package kratos

import (
	"log/slog"
	"time"

	kratosclientgo "github.com/ory/kratos-client-go"
)

type Session kratosclientgo.Session

func NewFromKratosClientSession(session *kratosclientgo.Session) *Session {
	s := Session(*session)
	return &s
}

func (s *Session) GetValueFromTraits(key string) string {
	if s == nil {
		slog.Info("Session is nil")
		return ""
	}
	traits := s.Identity.Traits.(map[string]interface{})
	if traits[key] == nil {
		return ""
	}
	email, ok := traits[key].(string)
	if !ok {
		email = ""
	}
	return email
}

// セッションがprivileged_session_max_age を過ぎているかどうかを返却する
func (s *Session) NeedLoginWhenPrivilegedAccess() bool {
	authenticateAt := s.AuthenticatedAt.In(locationJst)
	if authenticateAt.Before(time.Now().Add(-time.Second * privilegedAccessLimitMinutes)) {
		return true
	} else {
		return false
	}
}
