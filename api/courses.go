package api

import (
	"encoding/json"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type CourseResponse struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	CourseCode             string `json:"course_code"`
	Concluded              bool   `json:"concluded"`
	AccessRestrictedByDate *bool  `json:"access_restricted_by_date"`
	Term                   struct {
		ID int32 `json:"id"`
	} `json:"term"`
	Enrollments []struct {
		ComputedCurrentScore *float32 `json:"computed_current_score"`
	} `json:"enrollments"`
	Teachers []struct {
		DisplayName string `json:"display_name"`
	} `json:"teachers"`
}

func (c *Client) GetCourses() ([]CourseResponse, error) {
	url := config.BaseURL + "/api/v1/courses?include[]=teachers&include[]=total_scores&include[]=concluded&include[]=term&per_page=100"

	resp, err := c.Fetch(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out []CourseResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}
