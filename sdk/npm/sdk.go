package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://registry.npmjs.org"

type alternativeVersion struct {
	Repository string `json:"repository"`
}

type Version struct {
	Repository struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"repository"`
}

func FetchLatestVersion(ctx context.Context, cli *http.Client, lib string) (*Version, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%v/%v/latest", baseURL, lib), nil)
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
