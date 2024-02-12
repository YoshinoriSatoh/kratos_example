package kratos

import (
	"log/slog"
	"os"

	kratosclientgo "github.com/ory/kratos-client-go"
)

type Provider struct {
	d                  Dependencies
	kratosPublicClient *kratosclientgo.APIClient
}

type Dependencies struct {
}

type NewInput struct {
	KratosPublicEndpoint string
}

func New(i NewInput, d Dependencies) (*Provider, error) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	p := Provider{
		d: d,
	}
	kratosPublicConfigration := kratosclientgo.NewConfiguration()
	kratosPublicConfigration.Servers = []kratosclientgo.ServerConfiguration{{URL: i.KratosPublicEndpoint}}
	p.kratosPublicClient = kratosclientgo.NewAPIClient(kratosPublicConfigration)

	return &p, nil
}
