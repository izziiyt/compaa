package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "https://hub.docker.com/v2"

type Response struct {
	LastUpdated   time.Time `json:"last_updated"`
	TagLastPushed time.Time `json:"tag_last_pushed"`
}

func ReadTag(ctx context.Context, namespace, repository, tag string) (*Response, error) {
	cli := http.DefaultClient
	url := fmt.Sprintf("%s/namespaces/%s/repositories/%s/tags/%s", baseURL, namespace, repository, tag)
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
