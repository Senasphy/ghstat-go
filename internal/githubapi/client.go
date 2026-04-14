package githubapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
)

type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

func NewClient(token string) (graphql.Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing GitHub token: pass --token or set GITHUB_TOKEN")
	}

	httpClient := &http.Client{
		Transport: authTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}

	return graphql.NewClient("https://api.github.com/graphql", httpClient), nil
}
