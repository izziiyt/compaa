package gopkg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var githubURLRegexp = regexp.MustCompile(`https://github\.com/[\w-]+/[\w-]+`)

func GetGitHub(ctx context.Context, cli *http.Client, name string) (org, repo string, err error) {
	url := fmt.Sprintf("https://%v", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	res, err := cli.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		//nolint:errcheck
		io.Copy(io.Discard, res.Body)
		return
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	match := githubURLRegexp.FindString(string(b))
	tokens := strings.Split(match, "/")
	if len(tokens) > 4 {
		org = tokens[3]
		repo = tokens[4]
	}
	return
}
