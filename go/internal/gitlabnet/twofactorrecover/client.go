package twofactorrecover

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet"
)

type Client struct {
	config *config.Config
	client gitlabnet.GitlabClient
}

type Response struct {
	Success       bool     `json:"success"`
	RecoveryCodes []string `json:"recovery_codes"`
	Message       string   `json:"message"`
}

type RequestBody struct {
	KeyId string `json:"key_id"`
}

func NewClient(config *config.Config) (*Client, error) {
	client, err := gitlabnet.GetClient(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating http client: %v", err)
	}

	return &Client{config: config, client: client}, nil
}

func (c *Client) GetRecoveryCodes(gitlabKeyId string) ([]string, error) {
	values := RequestBody{KeyId: gitlabKeyId}
	response, err := c.client.Post("/two_factor_recovery_codes", values)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	parsedResponse, err := c.parseResponse(response)

	if err != nil {
		return nil, fmt.Errorf("Parsing failed")
	}

	if parsedResponse.Success {
		return parsedResponse.RecoveryCodes, nil
	} else {
		return nil, errors.New(parsedResponse.Message)
	}
}

func (c *Client) parseResponse(resp *http.Response) (*Response, error) {
	parsedResponse := &Response{}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, parsedResponse); err != nil {
		return nil, err
	} else {
		return parsedResponse, nil
	}
}
