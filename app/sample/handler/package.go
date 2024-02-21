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
	tmpl         *template.Template
	validate     *validator.Validate
	trans        ut.Translator
	cookieParams CookieParams
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
	loadTemplate()
	initValidator()
	cookieParams = i.CookieParams
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
