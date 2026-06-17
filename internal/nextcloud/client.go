package nextcloud

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	Host     string
	Username string
	Password string
	HTTP     *http.Client
}

func NewClient(host, username, password string) *Client {
	return &Client{
		Host:     strings.TrimRight(host, "/"),
		Username: username,
		Password: password,
		HTTP:     &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := c.Host + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("OCS-APIREQUEST", "true")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.HTTP.Do(req)
}

func (c *Client) DoJSON(method, path string, body io.Reader, result interface{}) error {
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	resp, err := c.doRequest(method, path, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

func (c *Client) DoXML(method, path string, body io.Reader, headers map[string]string, result interface{}) (int, error) {
	resp, err := c.doRequest(method, path, body, headers)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if result != nil && len(data) > 0 {
		return resp.StatusCode, xml.Unmarshal(data, result)
	}
	return resp.StatusCode, nil
}

func (c *Client) DoRaw(method, path string, body io.Reader, headers map[string]string) ([]byte, int, error) {
	resp, err := c.doRequest(method, path, body, headers)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}
