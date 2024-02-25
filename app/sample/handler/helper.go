package handler

import (
	"context"
	"kratos_example/kratos"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func getSession(ctx context.Context) *kratos.Session {
	session := ctx.Value("session")
	if session == nil {
		return nil
	}

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

func validationFieldErrors(err error) map[string]string {
	if err == nil {
		return map[string]string{}
	}

	fieldsErrors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		fieldsErrors[err.StructField()] = err.Translate(trans)
	}
	return fieldsErrors
}

func setCookieToResponseHeader(w http.ResponseWriter, cookies []string) {
	for _, cookie := range cookies {
		w.Header().Add("Set-Cookie", cookie)
	}
}

func redirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	if r.Header.Get("HX-Request") == "true" {
		slog.Info("HX-Redirect")
		w.Header().Set("HX-Redirect", redirectTo)
		// w.Header().Set("HX-Location", redirectTo)
		// w.WriteHeader(http.StatusSeeOther)
	} else {
		slog.Info("Redirect")
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}

func viewParameters(session *kratos.Session, r *http.Request, p map[string]any) map[string]any {
	params := p
	params["IsAuthenticated"] = isAuthenticated(session)
	params["Navbar"] = getNavbarviewParameters(session)
	params["CurrentPath"] = r.URL.Path
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
