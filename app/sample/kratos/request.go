package kratos

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type requestKratosInput struct {
	Method     string
	Path       string
	BodyBytes  []byte
	Cookie     string
	RemoteAddr string
}

type requestKratosOutput struct {
	BodyBytes  []byte
	Header     http.Header
	StatusCode int
}

func (p *Provider) requestKratosPublic(i requestKratosInput) (requestKratosOutput, error) {
	slog.Info(pkgVars.kratosPublicEndpoint)
	return requestKratos(pkgVars.kratosPublicEndpoint, i)
}

func (p *Provider) requestKratosAdmin(i requestKratosInput) (requestKratosOutput, error) {
	return requestKratos(pkgVars.kratosAdminEndpoint, i)
}

func requestKratos(endpoint string, i requestKratosInput) (requestKratosOutput, error) {
	slog.Info(fmt.Sprintf("%s%s", endpoint, i.Path))
	req, err := http.NewRequest(
		i.Method,
		fmt.Sprintf("%s%s", endpoint, i.Path),
		bytes.NewBuffer(i.BodyBytes))
	if err != nil {
		slog.Error("NewRequestError", "Error", err)
		return requestKratosOutput{}, err
	}
	req.Header.Set("Cookie", i.Cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("True-Client-IP", i.RemoteAddr)
	// req.Header.Set("X-Forwarded-For", i.RemoteAddr)
	slog.Info(fmt.Sprintf("%v", req))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("http error", "Error", err)
		return requestKratosOutput{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error())
		return requestKratosOutput{}, err
	}
	slog.Info(string(body))
	return requestKratosOutput{
		BodyBytes:  body,
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
	}, nil
}
