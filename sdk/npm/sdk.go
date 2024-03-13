package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type Version struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Main        string `json:"main"`
	Scripts     struct {
		Test string `json:"test"`
	} `json:"scripts"`
	Repository struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"repository"`
	Keywords []string `json:"keywords"`
	Author   struct {
		Name string `json:"name"`
	} `json:"author"`
	License string `json:"license"`
	Bugs    struct {
		Url string `json:"url"`
	} `json:"bugs"`
	Homepage        string `json:"homepage"`
	DevDependencies struct {
		Typescript string `json:"typescript"`
	} `json:"devDependencies"`
	Dependencies struct {
		BabelCore      string `json:"@babel/core"`
		BabelGenerator string `json:"@babel/generator"`
		BabelParser    string `json:"@babel/parser"`
		BabelTraverse  string `json:"@babel/traverse"`
		BabelTypes     string `json:"@babel/types"`
	} `json:"dependencies"`
	GitHead     string `json:"gitHead"`
	Id          string `json:"_id"`
	NodeVersion string `json:"_nodeVersion"`
	NpmVersion  string `json:"_npmVersion"`
	Dist        struct {
		Integrity    string `json:"integrity"`
		Shasum       string `json:"shasum"`
		Tarball      string `json:"tarball"`
		FileCount    int    `json:"fileCount"`
		UnpackedSize int    `json:"unpackedSize"`
		NpmSignature string `json:"npm-signature"`
		Signatures   []struct {
			Keyid string `json:"keyid"`
			Sig   string `json:"sig"`
		} `json:"signatures"`
	} `json:"dist"`
	NpmUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"_npmUser"`
	Directories struct {
	} `json:"directories"`
	Maintainers []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"maintainers"`
	NpmOperationalInternal struct {
		Host string `json:"host"`
		Tmp  string `json:"tmp"`
	} `json:"_npmOperationalInternal"`
	HasShrinkwrap bool `json:"_hasShrinkwrap"`
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
		return nil, err
	}

	return v, nil
}
