package handler

import (
	"net/http"
)

// Home画面（ログイン必須）レンダリング
type item struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

var items = []item{
	{
		Name:        "Item1",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item2",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
	{
		Name:        "Item3",
		Image:       "https://daisyui.com/images/stock/photo-1606107557195-0e29a4b5b4aa.jpg",
		Description: "Item1 Description",
		Link:        "/item/1",
	},
}

func (p *Provider) handleGetTop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	tmpl.ExecuteTemplate(w, "top/index.html", viewParameters(session, r, map[string]any{
		"Items": items,
	}))
}

func (p *Provider) handleGetItemDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := getSession(ctx)

	r.PathValue("id")

	tmpl.ExecuteTemplate(w, "item/detail.html", viewParameters(session, r, map[string]any{
		"Items": items,
	}))
}
