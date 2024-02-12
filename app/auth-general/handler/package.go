package handler

import "html/template"

var (
	generalEndpoint string
	tmpl            *template.Template
)

type InitInput struct {
	GeneralEndpoint string
}

func Init(i InitInput) {
	generalEndpoint = i.GeneralEndpoint

	tmpl = template.Must(template.New("").ParseGlob("templates/**/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/**/**/*.html"))
}
