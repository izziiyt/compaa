package gopkg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Client struct {
	Cli *http.Client
}

func NewClient(cli *http.Client) *Client {
	if cli == nil {
		cli = http.DefaultClient
	}
	return &Client{Cli: cli}
}

func (c *Client) GetGitHub(ctx context.Context, name string) (org, repo string, err error) {
	url := fmt.Sprintf("https://%v", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	res, err := c.Cli.Do(req)
	defer res.Body.Close()
	if err != nil {
		io.Copy(io.Discard, res.Body)
		return
	}
	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, res.Body)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`"https://github\.com/.*"`)
	match := re.FindString(string(b))
	tokens := strings.Split(match, "/")
	if len(tokens) > 5 {
		org = tokens[3]
		repo = tokens[4]
	}
	return
}
