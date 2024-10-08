package rubygem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://rubygems.org/api/v1/gems"

type Response struct {
	SourceCodeURI    string `json:"source_code_uri"`
	HomepageURI      string `json:"homepage_uri"`
	DocumentationURI string `json:"documentation_uri"`
	BugTrackerURI    string `json:"bug_tracker_uri"`
}

func GetGem(ctx context.Context, cli *http.Client, name string) (*Response, error) {
	url := fmt.Sprintf("%s/%s.json", baseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		//nolint:errcheck
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

	return r, nil
}
