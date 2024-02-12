package main

import (
	"html/template"
	"kratos_example/handler"
	"kratos_example/kratos"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	kratosProvider  *kratos.Provider
	handlerProvider *handler.Provider
)

func init() {
	// Set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	// Init packages
	kratos.Init(kratos.InitInput{
		PrivilegedAccessLimitMinutes: 10,
	})

	handler.Init(handler.InitInput{
		GeneralEndpoint: "http://localhost:3000",
	})

	// Create package providers with dependencies
	var err error

	kratosProvider, err = kratos.New(
		kratos.NewInput{
			KratosPublicEndpoint: "http://kratos-general:4433",
		},
		kratos.Dependencies{},
	)
	if err != nil {
		panic(err)
	}

	handlerProvider, err = handler.New(
		handler.NewInput{},
		handler.Dependencies{
			Kratos: kratosProvider,
		},
	)
	if err != nil {
		panic(err)
	}
}

func main() {
	e := gin.New()
	e.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/public/health"},
	}))
	e.Use(gin.Recovery())

	templateList := template.Must(template.New("").ParseGlob("templates/**/*.html"))
	templateList = template.Must(templateList.ParseGlob("templates/**/**/*.html"))

	e.SetHTMLTemplate(templateList)
	e.Static("/static", "./static")

	handlerProvider.Register(e)

	e.Run(":3000")
}
