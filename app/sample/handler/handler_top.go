package handler

import (
	"net/http"
)

var items = []item{
	{
		Name:        "Item1",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item1 Description",
		Link:        "/item/1",
		Price:       1000,
	},
	{
		Name:        "Item2",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item2 Description",
		Link:        "/item/2",
		Price:       1000,
	},
	{
		Name:        "Item3",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item3 Description",
		Link:        "/item/3",
		Price:       1000,
	},
	{
		Name:        "Item4",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item4 Description",
		Link:        "/item/4",
		Price:       1000,
	},
	{
		Name:        "Item5",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item5 Description",
		Link:        "/item/5",
		Price:       1000,
	},
	{
		Name:        "Item6",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item6 Description",
		Link:        "/item/6",
		Price:       1000,
	},
	{
		Name:        "Item7",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item7 Description",
		Link:        "/item/7",
		Price:       1000,
	},
	{
		Name:        "Item8",
		Image:       "http://localhost:3000/static/sample.png",
		Description: "Item8 Description",
		Link:        "/item/8",
		Price:       1000,
	},
}

func (p *Provider) handleGetTop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	pkgVars.tmpl.ExecuteTemplate(w, "top/index.html", viewParameters(session, r, map[string]any{
		"Items": items,
	}))
}
