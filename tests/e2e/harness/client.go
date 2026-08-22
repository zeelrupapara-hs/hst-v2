package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// Connection types accepted by the login routes.
const (
	ConnTrader  = 0
	ConnAdmin   = 32
	ConnManager = 33
)

// Response is the envelope every hst-server route answers with.
type Response struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
	Message string          `json:"message"`
}

// Token is what a login returns.
type Token struct {
	Login        int64  `json:"login"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionId    string `json:"session_id"`
}

// Client is one authenticated http session against hst-server.
type Client struct {
	Base  string
	Token Token
	http  *http.Client
}

// NewClient makes an unauthenticated client.
func NewClient(base string) *Client {
	return &Client{Base: base, http: &http.Client{Timeout: 15 * time.Second}}
}

// Login does a Basic login against path and keeps the token; returns the status.
func (c *Client) Login(path string, login int64, password string, connType int) (int, Response) {
	body, _ := json.Marshal(map[string]int{"connection_type": connType})
	req, _ := http.NewRequest("POST", c.Base+path, bytes.NewReader(body))
	req.SetBasicAuth(fmt.Sprint(login), password)
	req.Header.Set("Content-Type", "application/json")
	st, res := c.send(req)
	if st == 200 {
		_ = json.Unmarshal(res.Data, &c.Token)
	}
	return st, res
}

// Request sends one request and decodes data into out when given; for code that has no *testing.T.
func (c *Client) Request(method, path string, body, out any) (int, Response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, Response{}, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.Base+path, reader)
	if err != nil {
		return 0, Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token.AccessToken)
	}
	st, res := c.send(req)
	if st == 0 {
		return 0, res, fmt.Errorf("%s %s: %s", method, path, res.Error)
	}
	if out != nil && len(res.Data) > 0 {
		if err := json.Unmarshal(res.Data, out); err != nil {
			return st, res, fmt.Errorf("%s %s: decode data: %w", method, path, err)
		}
	}
	return st, res, nil
}

// Do is Request for a test: a transport error fails the test.
func (c *Client) Do(t *testing.T, method, path string, body, out any) (int, Response) {
	t.Helper()
	st, res, err := c.Request(method, path, body, out)
	if err != nil {
		t.Fatal(err)
	}
	return st, res
}

func (c *Client) send(req *http.Request) (int, Response) {
	var res Response
	resp, err := c.http.Do(req)
	if err != nil {
		res.Error = err.Error()
		return 0, res
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &res)
	}
	return resp.StatusCode, res
}

// Get is Do with GET.
func (c *Client) Get(t *testing.T, path string, out any) (int, Response) {
	t.Helper()
	return c.Do(t, "GET", path, nil, out)
}

// Post is Do with POST.
func (c *Client) Post(t *testing.T, path string, body, out any) (int, Response) {
	t.Helper()
	return c.Do(t, "POST", path, body, out)
}

// Patch is Do with PATCH.
func (c *Client) Patch(t *testing.T, path string, body, out any) (int, Response) {
	t.Helper()
	return c.Do(t, "PATCH", path, body, out)
}

// Put is Do with PUT.
func (c *Client) Put(t *testing.T, path string, body, out any) (int, Response) {
	t.Helper()
	return c.Do(t, "PUT", path, body, out)
}

// Delete is Do with DELETE.
func (c *Client) Delete(t *testing.T, path string) (int, Response) {
	t.Helper()
	return c.Do(t, "DELETE", path, nil, nil)
}

// M is a short json object literal.
type M = map[string]any
