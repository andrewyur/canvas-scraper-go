package api

import (
	"encoding/json"
)

type FileResponse struct {
	Url *string `json:"url"`
}

func (c *Client) GetFileDownloadUrl(url string) (*FileResponse, error) {
	resp, err := c.Fetch(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
