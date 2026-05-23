package haybtech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.haybtech.com/v1"
	DefaultTimeout = 15 * time.Second
)

type Client struct {
	SecretKey string
	BaseURL   string
	HTTPClient *http.Client

	Payments *PaymentsResource
}

func NewClient(secretKey string) (*Client, error) {
	if !strings.HasPrefix(secretKey, "sk_") {
		return nil, fmt.Errorf("invalid secret key: must start with sk_")
	}

	// CRLF injection check
	if strings.ContainsAny(secretKey, "\r\n") {
		return nil, fmt.Errorf("invalid secret key: contains forbidden characters")
	}

	c := &Client{
		SecretKey: secretKey,
		BaseURL:   DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	c.Payments = &PaymentsResource{client: c}
	return c, nil
}

// Security: Mask secret key when printing the client
func (c *Client) String() string {
	masked := "sk_..." + c.SecretKey[len(c.SecretKey)-4:]
	return fmt.Sprintf("HayBTechClient{BaseURL: %s, SecretKey: %s}", c.BaseURL, masked)
}

func (c *Client) request(method, path string, body interface{}, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(c.BaseURL, "/"), strings.TrimPrefix(path, "/"))
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "HayBTech-Go-SDK/1.0.0")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type PaymentsResource struct {
	client *Client
}

func (r *PaymentsResource) Create(params map[string]interface{}) (map[string]interface{}, error) {
	resp, err := r.client.request("POST", "/payments", params, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
