package gopkg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func GetRepoFromCustomDomain(ctx context.Context, name string) (org, repo string, err error) {
	url := fmt.Sprintf("https://%v", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
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
	re := regexp.MustCompile(`<meta name="go-import" content=".*">`)
	match := re.FindString(string(b))
	if match == "" {
		err = fmt.Errorf("no go-import meta tag found")
		return
	}
	match = strings.TrimPrefix(match, "<meta name=\"go-import\" content=\"")
	match = strings.TrimSuffix(match, "\">")
	repoUrl := strings.Split(match, " ")[2]
	tokens := strings.Split(repoUrl, "/")
	if len(tokens) >= 5 {
		org = tokens[3]
		repo = tokens[4]
	} else {
		err = fmt.Errorf("unsupported regisitry %v", repoUrl)
	}
	return
}
