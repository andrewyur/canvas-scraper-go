package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andrewyur/canvas-scraper-go/requests"
)

type file struct {
	Url string `json:"url"`
}

func getFileDownloadUrl(client *http.Client, token, url string) (string, error) {
	resp, err := requests.Fetch(client, token, url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var f file
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return "", err
	}

	if len(f.Url) == 0 {
		return "", errors.New("no url returned")
	}

	return f.Url, nil
}
