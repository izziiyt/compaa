package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	cli     *http.Client
	BaseURL string
}

func NewClient(cli *http.Client) *Client {
	c := &Client{
		cli:     cli,
		BaseURL: "https://registry.npmjs.org",
	}
	if c.cli == nil {
		c.cli = http.DefaultClient
	}
	return c
}

type alternativeVersion struct {
	Repository string `json:"repository"`
}

type Version struct {
	Repository struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"repository"`
}

func (c *Client) FetchLatestVersion(ctx context.Context, lib string) (*Version, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%v/%v/latest", c.BaseURL, lib), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.cli.Do(req)
	defer res.Body.Close()
	if err != nil {
		io.Copy(io.Discard, res.Body)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, res.Body)
		return nil, fmt.Errorf("")
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	v := &Version{}
	if err := json.Unmarshal(b, v); err != nil {
		v1 := &alternativeVersion{}
		if err := json.Unmarshal(b, v1); err != nil {
			return nil, err
		}
		v.Repository.Url = v1.Repository
	}

	if !strings.Contains(v.Repository.Url, "github.com") {
		return nil, fmt.Errorf("unsupported registry %v", v.Repository.Url)
	}

	// Note: https://registry.npmjs.org/eslint-config-next/latest
	// does not have "repository.type" field
	v.Repository.Type = "git"

	return v, nil
}
