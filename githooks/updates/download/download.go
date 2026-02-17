package download

import (
	"net/http"
	"os"

	cm "github.com/gabyx/githooks/githooks/common"
)

// GetFile downloads a file from a `url`.
// Response body needs to be closed by caller.
func GetFile(url string) (response *http.Response, err error) {
	// Get the response bytes from the url
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		err = e

		return
	}

	if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		return nil, cm.ErrorF("Download of '%s' failed with status: '%v'.",
			url, response.StatusCode)
	}

	return
}
