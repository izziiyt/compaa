package eol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "https://endoflife.date/api"

type cycleDetail struct {
	ReleaseDate       string      `json:"releaseDate"`
	EOL               interface{} `json:"eol"`
	Latest            string      `json:"latest"`
	LatestReleaseDate string      `json:"latestReleaseDate"`
	LTS               bool        `json:"lts"`
}

type CycleDetail struct {
	ReleaseDate       time.Time
	EOLDate           time.Time
	EOL               bool
	Latest            string
	LatestReleaseDate time.Time
	LTS               bool
}

func NewCycleDetail(cd *cycleDetail) (CD *CycleDetail, err error) {
	CD = &CycleDetail{
		Latest: cd.Latest,
		LTS:    cd.LTS,
	}
	CD.ReleaseDate, err = time.Parse(time.DateOnly, cd.ReleaseDate)
	if err != nil {
		CD = nil
		return
	}
	CD.LatestReleaseDate, err = time.Parse(time.DateOnly, cd.LatestReleaseDate)
	if err != nil {
		CD = nil
		return
	}
	switch v := cd.EOL.(type) {
	case bool:
		CD.EOL = false
	case string:
		CD.EOL = true
		CD.EOLDate, err = time.Parse(time.DateOnly, v)
		if err != nil {
			CD = nil
			return
		}
		if CD.EOLDate.Before(time.Now()) {
			CD.EOL = true
		}
	default:
		CD = nil
		err = fmt.Errorf("unexpected type")
		return
	}
	return
}

func SingleCycleDetail(ctx context.Context, product, cycle string) (*CycleDetail, error) {
	url := fmt.Sprintf("%s/%s/%s.json", baseURL, product, cycle)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	cli := http.DefaultClient
	res, err := cli.Do(req)
	defer res.Body.Close()
	if err != nil {
		io.Copy(io.Discard, res.Body)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, res.Body)
		return nil, fmt.Errorf("something wrong with accesing endoflife.date code:%v", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	cd := &cycleDetail{}
	if err := json.Unmarshal(b, cd); err != nil {
		return nil, err
	}
	return NewCycleDetail(cd)
}
