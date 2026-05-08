package githubapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
)

type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	resp, err := base.RoundTrip(clone)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	return resp, nil
}

func NewClient(token string) (graphql.Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing GitHub token: pass --token or set GITHUB_TOKEN")
	}
	if strings.ContainsAny(token, "\r\n") {
		return nil, errors.New("invalid GitHub token: contains newline characters")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: authTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}

	return graphql.NewClient("https://api.github.com/graphql", httpClient), nil
}
