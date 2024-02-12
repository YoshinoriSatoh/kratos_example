package handler

import (
	"context"
	"fmt"
	"kratos_example/kratos"
	"log/slog"
)

func getSession(ctx context.Context) *kratos.Session {
	session := ctx.Value("session")
	if session == nil {
		return nil
	}
	slog.Info(fmt.Sprintf("%v", session))

	kratosSession, _ := session.(*kratos.Session)
	return kratosSession
}

func isAuthenticated(session *kratos.Session) bool {
	if session != nil {
		return true
	} else {
		return false
	}
}

func existsTraitsFieldsNotFilledIn(session *kratos.Session) bool {
	if session.GetValueFromTraits("email") == "" ||
		session.GetValueFromTraits("nickname") == "" ||
		session.GetValueFromTraits("birthdate") == "" {
		return true
	} else {
		return false
	}
}
