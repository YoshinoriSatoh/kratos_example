package handler

import (
	"net/http"
)

var items = []item{
	{
		Name:        "Item1",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
		Price:       1000,
	},
	{
		Name:        "Item2",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item2 Description",
		Link:        "/item/2",
		Price:       1000,
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item3 Description",
		Link:        "/item/3",
		Price:       1000,
	},
	{
		Name:        "Item4",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item4 Description",
		Link:        "/item/4",
		Price:       1000,
	},
	{
		Name:        "Item5",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item5 Description",
		Link:        "/item/5",
		Price:       1000,
	},
	{
		Name:        "Item6",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item6 Description",
		Link:        "/item/6",
		Price:       1000,
	},
	{
		Name:        "Item7",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item7 Description",
		Link:        "/item/7",
		Price:       1000,
	},
	{
		Name:        "Item8",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item8 Description",
		Link:        "/item/8",
		Price:       1000,
	},
}

func (p *Provider) handleGetTop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	tmpl.ExecuteTemplate(w, "top/index.html", viewParameters(session, r, map[string]any{
		"Items": items,
	}))
}
