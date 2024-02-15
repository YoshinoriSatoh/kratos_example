package handler

import (
	"html/template"
	"reflect"

	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

var (
	tmpl          *template.Template
	validate      *validator.Validate
	trans         ut.Translator
	routePaths    _routePaths
	templatePaths _templatePaths
	cookieParams  CookieParams
)

type CookieParams struct {
	SessionCookieName string
	Path              string
	Domain            string
	Secure            bool
}

type InitInput struct {
	CookieParams CookieParams
}

func Init(i InitInput) {
	initRoutePaths()
	initTemplatePaths()
	loadTemplate()
	initValidator()
	cookieParams = i.CookieParams
}

type _routePaths struct {
	AuthRegistration      string
	AuthVerification      string
	AuthVerificationEmail string
	AuthVerificationCode  string
	AuthLogin             string
	AuthLogout            string
	AuthRecovery          string
	AuthRecoveryEmail     string
	AuthRecoveryCode      string
	MyPassword            string
	MyProfile             string
	MyProfileEdit         string
	MyProfileForm         string
	Top                   string
	Item                  string
}

func initRoutePaths() {
	routePaths = _routePaths{
		AuthRegistration:      "/auth/registration",
		AuthVerification:      "/auth/verification",
		AuthVerificationEmail: "/auth/verification/email",
		AuthVerificationCode:  "/auth/verification/code",
		AuthLogin:             "/auth/login",
		AuthLogout:            "/auth/logout",
		AuthRecovery:          "/auth/recovery",
		AuthRecoveryEmail:     "/auth/recovery/email",
		AuthRecoveryCode:      "/auth/recovery/code",
		MyPassword:            "/my/password",
		MyProfile:             "/my/profile",
		MyProfileEdit:         "/my/profile/edit",
		MyProfileForm:         "/my/profile/form",
		Top:                   "/",
		Item:                  "/item",
	}
}

type _templatePaths struct {
	Layout_Header              string
	Layout_Navbar              string
	Layout_Footer              string
	AuthRegistrationIndex      string
	AuthRegistration_Form      string
	AuthVerificationIndex      string
	AuthVerificationCode       string
	AuthVerification_EmailForm string
	AuthVerification_CodeForm  string
	AuthLoginIndex             string
	AuthLogin_Form             string
	AuthRecoveryIndex          string
	AuthRecovery_EmalForm      string
	AuthRecovery_CodeForm      string
	MyPasswordIndex            string
	MyPassword_Form            string
	MyProfileIndex             string
	MyProfileEdit              string
	MyProfile_Form             string
	MyProfile_Edit             string
	TopIndex                   string
}

func initTemplatePaths() {
	templatePaths = _templatePaths{
		Layout_Header:              "layout/_header.html",
		Layout_Navbar:              "layout/_navbar.html",
		Layout_Footer:              "layout/_footer.html",
		AuthRegistrationIndex:      "auth/registration/index.html",
		AuthRegistration_Form:      "auth/registration/_form.html",
		AuthVerificationIndex:      "auth/verification/index.html",
		AuthVerificationCode:       "auth/verification/code.html",
		AuthVerification_EmailForm: "auth/verification/_email_form.html",
		AuthVerification_CodeForm:  "auth/verification/_code_form.html",
		AuthLoginIndex:             "auth/login/index.html",
		AuthLogin_Form:             "auth/login/_form.html",
		AuthRecoveryIndex:          "auth/recovery/index.html",
		AuthRecovery_EmalForm:      "auth/recovery/_email_form.html",
		AuthRecovery_CodeForm:      "auth/recovery/_code_form.html",
		MyPasswordIndex:            "my/password/index.html",
		MyPassword_Form:            "my/password/_form.html",
		MyProfileIndex:             "my/profile/index.html",
		MyProfileEdit:              "my/profile/edit.html",
		MyProfile_Form:             "my/profile/_form.html",
		MyProfile_Edit:             "my/profile/_edit.html",
		TopIndex:                   "top/index.html",
	}
}

func loadTemplate() {
	tmpl = template.Must(template.New("").ParseGlob("templates/**/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/**/**/*.html"))
}

func initValidator() {
	ja := ja.New()
	uni := ut.New(ja)
	trans, _ = uni.GetTranslator("ja")

	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		fieldName := field.Tag.Get("ja")
		if fieldName == "-" {
			return ""
		}
		return fieldName
	})
	err := ja_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(err)
	}
}
