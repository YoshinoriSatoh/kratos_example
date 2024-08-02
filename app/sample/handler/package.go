package handler

import (
	"html/template"
	"reflect"
	"time"

	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

var pkgVars packageVariables

type packageVariables struct {
	tmpl            *template.Template
	validate        *validator.Validate
	trans           ut.Translator
	cookieParams    CookieParams
	birthdateFormat string
}

type CookieParams struct {
	SessionCookieName string
	Path              string
	Domain            string
	Secure            bool
}

type InitInput struct {
	CookieParams    CookieParams
	BirthdateFormat string
}

func Init(i InitInput) {
	loadTemplate()
	initValidator()
	pkgVars.cookieParams = i.CookieParams
	pkgVars.birthdateFormat = i.BirthdateFormat
}

func loadTemplate() {
	pkgVars.tmpl = template.Must(template.New("").ParseGlob("templates/**/*.html"))
	pkgVars.tmpl = template.Must(pkgVars.tmpl.ParseGlob("templates/**/**/*.html"))
}

func initValidator() {
	ja := ja.New()
	uni := ut.New(ja)
	pkgVars.trans, _ = uni.GetTranslator("ja")

	pkgVars.validate = validator.New(validator.WithRequiredStructEnabled())
	pkgVars.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		fieldName := field.Tag.Get("ja")
		if fieldName == "-" {
			return ""
		}
		return fieldName
	})
	err := ja_translations.RegisterDefaultTranslations(pkgVars.validate, pkgVars.trans)
	if err != nil {
		panic(err)
	}
	pkgVars.validate.RegisterValidation("birthdate", validateBirthdate)
}

func validateBirthdate(fl validator.FieldLevel) bool {
	_, err := time.Parse(pkgVars.birthdateFormat, fl.Field().String())
	return err == nil
}
