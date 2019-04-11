package gitlabnet

import (
	"context"
	"net"
	"net/http"
	"strings"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
)

const (
	socketBaseUrl      = "http://unix"
	UnixSocketProtocol = "http+unix://"
	HttpProtocol       = "http://"
)

type GitlabHttpClient struct {
	httpClient *http.Client
	config     *config.Config
	host       string
}

func buildSocketClient(config *config.Config) *GitlabHttpClient {
	path := strings.TrimPrefix(config.GitlabUrl, UnixSocketProtocol)
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "unix", path)
		},
	}

	return buildClient(config, transport, socketBaseUrl)
}

func buildHttpClient(config *config.Config) *GitlabHttpClient {
	return buildClient(config, &http.Transport{}, config.GitlabUrl)
}

func buildClient(config *config.Config, transport *http.Transport, host string) *GitlabHttpClient {
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.HttpSettings.ReadTimeout(),
	}

	return &GitlabHttpClient{httpClient: httpClient, config: config, host: host}
}

func (c *GitlabHttpClient) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", path, nil)
}

func (c *GitlabHttpClient) Post(path string, data interface{}) (*http.Response, error) {
	return c.doRequest("POST", path, data)
}

func (c *GitlabHttpClient) doRequest(method, path string, data interface{}) (*http.Response, error) {
	request, err := newRequest(method, c.host, path, data)
	if err != nil {
		return nil, err
	}

	user, password := c.config.HttpSettings.User, c.config.HttpSettings.Password
	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	}

	return doRequest(c.httpClient, c.config, request)
}
