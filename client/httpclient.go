package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/gitlab-org/labkit/correlation"
)

const (
	socketBaseUrl             = "http://unix"
	unixSocketProtocol        = "http+unix://"
	httpProtocol              = "http://"
	httpsProtocol             = "https://"
	defaultReadTimeoutSeconds = 300
)

type HttpClient struct {
	*http.Client
	Host string
}

type httpClientCfg struct {
	keyPath, certPath string
}

type HTTPClientOpt func(*httpClientCfg)

func WithClientCert(certPath, keyPath string) HTTPClientOpt {
	return func(hcc *httpClientCfg) {
		hcc.keyPath = keyPath
		hcc.certPath = certPath
	}
}

func NewHTTPClient(gitlabURL, gitlabRelativeURLRoot, caFile, caPath string, selfSignedCert bool, readTimeoutSeconds uint64, opts ...HTTPClientOpt) *HttpClient {
	httpClientCfg := &httpClientCfg{}

	for _, opt := range opts {
		opt(httpClientCfg)
	}

	var transport *http.Transport
	var host string
	if strings.HasPrefix(gitlabURL, unixSocketProtocol) {
		transport, host = buildSocketTransport(gitlabURL, gitlabRelativeURLRoot)
	} else if strings.HasPrefix(gitlabURL, httpProtocol) {
		transport, host = buildHttpTransport(gitlabURL)
	} else if strings.HasPrefix(gitlabURL, httpsProtocol) {
		transport, host = buildHttpsTransport(httpClientCfg.certPath, httpClientCfg.keyPath, caFile, caPath, selfSignedCert, gitlabURL)
	} else {
		return nil
	}

	c := &http.Client{
		Transport: correlation.NewInstrumentedRoundTripper(transport),
		Timeout:   readTimeout(readTimeoutSeconds),
	}

	client := &HttpClient{Client: c, Host: host}

	return client
}

func buildSocketTransport(gitlabURL, gitlabRelativeURLRoot string) (*http.Transport, string) {
	socketPath := strings.TrimPrefix(gitlabURL, unixSocketProtocol)

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}

	host := socketBaseUrl
	gitlabRelativeURLRoot = strings.Trim(gitlabRelativeURLRoot, "/")
	if gitlabRelativeURLRoot != "" {
		host = host + "/" + gitlabRelativeURLRoot
	}

	return transport, host
}

func buildHttpsTransport(certPath, keyPath, caFile, caPath string, selfSignedCert bool, gitlabURL string) (*http.Transport, string) {
	certPool, err := x509.SystemCertPool()

	if err != nil {
		certPool = x509.NewCertPool()
	}

	var clientCert tls.Certificate
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(keyPath, certPath)
		if err != nil {
			// TODO: log the error or return to caller
		}
		clientCert = cert
	}

	if caFile != "" {
		addCertToPool(certPool, caFile)
	}

	if caPath != "" {
		fis, _ := ioutil.ReadDir(caPath)
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}

			addCertToPool(certPool, filepath.Join(caPath, fi.Name()))
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			RootCAs:            certPool,
			InsecureSkipVerify: selfSignedCert,
		},
	}

	return transport, gitlabURL
}

func addCertToPool(certPool *x509.CertPool, fileName string) {
	cert, err := ioutil.ReadFile(fileName)
	if err == nil {
		certPool.AppendCertsFromPEM(cert)
	}
}

func buildHttpTransport(gitlabURL string) (*http.Transport, string) {
	return &http.Transport{}, gitlabURL
}

func readTimeout(timeoutSeconds uint64) time.Duration {
	if timeoutSeconds == 0 {
		timeoutSeconds = defaultReadTimeoutSeconds
	}

	return time.Duration(timeoutSeconds) * time.Second
}
