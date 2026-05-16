package api

import (
	"encoding/json"
	"strconv"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type ModuleResponse struct {
	Name  string `json:"name"`
	Items []struct {
		Url  string `json:"url"`
		Type string `json:"type"`
	} `json:"items"`
}

func (c *Client) GetModules(courseId int) ([]ModuleResponse, error) {
	url := config.BaseURL + "/api/v1/courses/" + strconv.Itoa(courseId) + "/modules?include[]=items&include[]=content_details&per_page=30"

	resp, err := c.Fetch(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out []ModuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}
