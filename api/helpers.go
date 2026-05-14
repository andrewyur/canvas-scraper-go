package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func fetchJSON(client *http.Client, token, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "canvas_session",
		Value: token,
	})

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
	}

	return resp.Body, nil
}

type file struct {
	Url string `json:"url"`
}

func getFileDownloadUrl(client *http.Client, token, url string) (string, error) {
	body, err := fetchJSON(client, token, url)
	if err != nil {
		return "", err
	}

	defer body.Close()

	var f file
	if err := json.NewDecoder(body).Decode(&f); err != nil {
		return "", err
	}

	return f.Url, nil
}
