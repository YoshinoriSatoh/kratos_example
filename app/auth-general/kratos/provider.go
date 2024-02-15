package kratos

import (
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
	p := Provider{
		d: d,
	}
	kratosPublicConfigration := kratosclientgo.NewConfiguration()
	kratosPublicConfigration.Servers = []kratosclientgo.ServerConfiguration{{URL: i.KratosPublicEndpoint}}
	p.kratosPublicClient = kratosclientgo.NewAPIClient(kratosPublicConfigration)

	return &p, nil
}
