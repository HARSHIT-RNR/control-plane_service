package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client implements OPA policy evaluation
type Client struct {
	baseURL    string
	policyPath string
	httpClient *http.Client
}

// NewClient creates a new OPA client
func NewClient(baseURL, policyPath string) *Client {
	return &Client{
		baseURL:    baseURL,
		policyPath: policyPath,
		httpClient: &http.Client{},
	}
}

// OPARequest represents a request to OPA
type OPARequest struct {
	Input map[string]interface{} `json:"input"`
}

// OPAResponse represents OPA's response
type OPAResponse struct {
	Result struct {
		Allow  bool   `json:"allow"`
		Reason string `json:"reason,omitempty"`
	} `json:"result"`
}

// Evaluate evaluates a policy decision with OPA
func (c *Client) Evaluate(ctx context.Context, input map[string]interface{}) (bool, string, error) {
	// Construct OPA API URL
	url := fmt.Sprintf("%s/v1/data/%s", c.baseURL, c.policyPath)

	// Create request payload
	request := OPARequest{
		Input: input,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return false, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var opaResponse OPAResponse
	if err := json.Unmarshal(body, &opaResponse); err != nil {
		return false, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return opaResponse.Result.Allow, opaResponse.Result.Reason, nil
}
