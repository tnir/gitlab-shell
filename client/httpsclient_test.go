package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/client/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/internal/testhelper"
)

func TestSuccessfulRequests(t *testing.T) {
	testCases := []struct {
		desc              string
		caFile, caPath    string
		selfSigned        bool
		certPath, keyPath string // used for TLS client certs
	}{
		{
			desc:   "Valid CaFile",
			caFile: path.Join(testhelper.TestRoot, "certs/valid/server.crt"),
		},
		{
			desc:   "Valid CaPath",
			caPath: path.Join(testhelper.TestRoot, "certs/valid"),
		},
		{
			desc:       "Self signed cert option enabled",
			selfSigned: true,
		},
		{
			desc:       "Invalid cert with self signed cert option enabled",
			caFile:     path.Join(testhelper.TestRoot, "certs/valid/server.crt"),
			selfSigned: true,
		},
		{
			desc:   "CA signed client cert",
			caFile: path.Join(testhelper.TestRoot, "certs/valid/server.crt"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			client, cleanup := setupWithRequests(t, tc.caFile, tc.caPath, "", "", tc.selfSigned)
			defer cleanup()

			response, err := client.Get(context.Background(), "/hello")
			require.NoError(t, err)
			require.NotNil(t, response)

			defer response.Body.Close()

			responseBody, err := ioutil.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, string(responseBody), "Hello")
		})
	}
}

func TestFailedRequests(t *testing.T) {
	testCases := []struct {
		desc   string
		caFile string
		caPath string
	}{
		{
			desc:   "Invalid CaFile",
			caFile: path.Join(testhelper.TestRoot, "certs/invalid/server.crt"),
		},
		{
			desc:   "Invalid CaPath",
			caPath: path.Join(testhelper.TestRoot, "certs/invalid"),
		},
		{
			desc: "Empty config",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			client, cleanup := setupWithRequests(t, tc.caFile, tc.caPath, "", "", false)
			defer cleanup()

			_, err := client.Get(context.Background(), "/hello")
			require.Error(t, err)

			assert.Equal(t, err.Error(), "Internal API unreachable")
		})
	}
}

func assertTLSClientCert(t *testing.T, tlsCS *tls.ConnectionState, certPath string) {
	if certPath == "" {
		return
	}

	raw, err := ioutil.ReadFile(certPath)
	require.NoError(t, err)

	pemBlock, _ := pem.Decode(raw)
	require.NotNil(t, pemBlock)

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	require.NoError(t, err)

	found := false
	for _, chain := range tlsCS.VerifiedChains {
		entity := chain[0]
		if reflect.DeepEqual(entity, cert) {
			found = true
		}
	}
	require.True(t, found)
}

func setupWithRequests(t *testing.T, caFile, caPath, certPath, keyPath string, selfSigned bool) (*GitlabNetClient, func()) {
	testDirCleanup, err := testhelper.PrepareTestRootDir()
	require.NoError(t, err)
	defer testDirCleanup()

	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/hello",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				assertTLSClientCert(t, r.TLS, certPath)
				require.Equal(t, http.MethodGet, r.Method)

				fmt.Fprint(w, "Hello")
			},
		},
	}

	url, cleanup := testserver.StartHttpsServer(t, requests)

	var opts []HTTPClientOpt
	if keyPath != "" && certPath != "" {
		opts = append(opts, WithClientCert(certPath, keyPath))
	}

	httpClient := NewHTTPClient(url, "", caFile, caPath, selfSigned, 1, opts...)

	client, err := NewGitlabNetClient("", "", "", httpClient)
	require.NoError(t, err)

	return client, cleanup
}
