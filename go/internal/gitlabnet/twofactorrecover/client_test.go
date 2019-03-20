package twofactorrecover

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"
)

var (
	testConfig *config.Config
	requests   []testserver.TestRequestHandler
)

func initialize(t *testing.T) {
	testConfig = &config.Config{GitlabUrl: "http+unix://" + testserver.TestSocket}
	requests = []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/two_factor_recovery_codes",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var requestBody *RequestBody
				json.Unmarshal(b, &requestBody)

				switch requestBody.KeyId {
				case "0":
					body := map[string]interface{}{
						"success":        true,
						"recovery_codes": [2]string{"recovery 1", "codes 1"},
					}
					json.NewEncoder(w).Encode(body)
				case "1":
					body := map[string]interface{}{
						"success": false,
						"message": "missing user",
					}
					json.NewEncoder(w).Encode(body)
				case "2":
					w.WriteHeader(http.StatusForbidden)
					body := &gitlabnet.ErrorResponse{
						Message: "Not allowed!",
					}
					json.NewEncoder(w).Encode(body)
				case "3":
					w.Write([]byte("{ \"message\": \"broken json!\""))
				case "4":
					w.WriteHeader(http.StatusForbidden)
				}
			},
		},
	}
}

func TestGetRecoveryCodes(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	result, err := client.GetRecoveryCodes("0")
	assert.NoError(t, err)
	assert.Equal(t, []string{"recovery 1", "codes 1"}, result)
}

func TestMissingUser(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	_, err := client.GetRecoveryCodes("1")
	assert.Equal(t, "missing user", err.Error())
}

func TestErrorResponses(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	testCases := []struct {
		desc          string
		fakeId        string
		expectedError string
	}{
		{
			desc:          "A response with an error message",
			fakeId:        "2",
			expectedError: "Not allowed!",
		},
		{
			desc:          "A response with bad JSON",
			fakeId:        "3",
			expectedError: "Parsing failed",
		},
		{
			desc:          "An error response without message",
			fakeId:        "4",
			expectedError: "Internal API error (403)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			resp, err := client.GetRecoveryCodes(tc.fakeId)

			assert.EqualError(t, err, tc.expectedError)
			assert.Nil(t, resp)
		})
	}
}

func setup(t *testing.T) (*Client, func()) {
	initialize(t)
	cleanup, err := testserver.StartSocketHttpServer(requests)
	require.NoError(t, err)

	client, err := NewClient(testConfig)
	require.NoError(t, err)

	return client, cleanup
}
