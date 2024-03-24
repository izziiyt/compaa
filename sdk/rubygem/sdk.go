package rubygem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var baseURL = "https://rubygems.org/api/v1/gems"

var githubURLRegexp = regexp.MustCompile(`https://github\.com/[\w-]+/[\w-]+`)

type Response struct {
	SourceCodeURI string `json:"source_code_uri"`
	// MetaData      struct {
	// 	HomepageURI  string `json:"homepage_uri"`
	// 	ChangeLogURI string `json:"changelog_uri"`
	// }
	HomepageURI      string `json:"homepage_uri"`
	DocumentationURI string `json:"documentation_uri"`
	BugTrackerURI    string `json:"bug_tracker_uri"`
}

func GetGem(ctx context.Context, name string) (*Response, error) {
	cli := http.DefaultClient
	url := fmt.Sprintf("%s/%s.json", baseURL, name)
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

	return r, nil
}
