package pypi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const baseURL = "https://pypi.org/pypi"

var githubURLRegexp = regexp.MustCompile(`https://github\.com/[\w-]+/[\w-]+`)

type Response struct {
	Info struct {
		Description string `json:"description"`
	} `json:"info"`
	RepositoryURL string
}

func GetPackage(ctx context.Context, name string) (*Response, error) {
	cli := http.DefaultClient
	url := fmt.Sprintf("%s/%s/json", baseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := cli.Do(req)
	defer res.Body.Close()
	if err != nil {
		io.Copy(io.Discard, res.Body)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, res.Body)
		return nil, fmt.Errorf("something wrong with accesing :%v %v", url, res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := &Response{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}

	match := githubURLRegexp.FindString(string(b))
	if match == "" {
		return r, fmt.Errorf("github repository url not found")
	}
	r.RepositoryURL = match
	return r, nil
}
