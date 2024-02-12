package handler

import (
	"kratos_example/kratos"

	"github.com/gin-gonic/gin"
)

func getSession(c *gin.Context) *kratos.Session {
	session, exists := c.Get("session")
	if !exists || session == nil {
		return nil
	}
	return session.(*kratos.Session)
}

func isAuthenticated(c *gin.Context) bool {
	session := getSession(c)
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
