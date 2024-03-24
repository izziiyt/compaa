package gcrio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"
)

const baseURL = "https://gcr.io/v2"

type _response struct {
	Manifest map[string]*Response `json:"manifest"`
}

type Response struct {
	Tag            []string `json:"tag"`
	TimeUploadedMS string   `json:"timeUploadedMs"`
	Uploaded       time.Time
}

func ReadTag(ctx context.Context, namespace, repository, tag string) (*Response, error) {
	cli := http.DefaultClient
	url := fmt.Sprintf("%s/%s/%s/tags/list", baseURL, namespace, repository)
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
		io.Copy(io.Discard, res.Body)
		return nil, fmt.Errorf("something wrong with accesing :%v %v", url, res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := &_response{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}

	for _, v := range r.Manifest {
		if slices.Contains(v.Tag, tag) {
			i, err := strconv.Atoi(v.TimeUploadedMS)
			if err != nil {
				return nil, err
			}
			return &Response{
				Tag:            v.Tag,
				TimeUploadedMS: v.TimeUploadedMS,
				Uploaded:       time.Unix(0, int64(i)*int64(time.Millisecond)),
			}, nil
		}
	}

	return nil, fmt.Errorf("tag %v not found", tag)
}
