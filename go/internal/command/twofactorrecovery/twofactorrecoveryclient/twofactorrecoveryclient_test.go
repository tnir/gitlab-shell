package twofactorrecoveryclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"
)

var (
	testConfig *config.Config
	requests   []testserver.TestRequestHandler
)

func init() {
	testConfig = &config.Config{GitlabUrl: "http+unix://" + testserver.TestSocket}
	requests = []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/two_factor_recovery_codes",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("key_id") == "1" {
					body := map[string]interface{}{
						"success":        true,
						"recovery_codes": [2]string{"recovery 1", "codes 1"},
					}
					json.NewEncoder(w).Encode(body)
				} else if r.URL.Query().Get("key_id") == "0" {
					body := map[string]interface{}{
						"success": false,
						"message": "missing user",
					}
					json.NewEncoder(w).Encode(body)
				} else {
					fmt.Fprint(w, "null")
				}
			},
		},
	}
}

func TestGetRecoveryCodes(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	result, err := client.GetRecoveryCodes("1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"recovery 1", "codes 1"}, result)
}

func TestMissingUser(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	_, err := client.GetRecoveryCodes("0")
	assert.Equal(t, "missing user", err.Error())
}

func setup(t *testing.T) (*Client, func()) {
	cleanup, err := testserver.StartSocketHttpServer(requests)
	require.NoError(t, err)

	client, err := GetClient(testConfig)
	require.NoError(t, err)

	return client, cleanup
}
