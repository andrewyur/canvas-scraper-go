package api

import (
	"encoding/json"
	"strconv"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type AssignmentResponse struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	PointsPossible  float32  `json:"points_possible"`
	SubmissionTypes []string `json:"submission_types"`
	Rubric          []struct {
		ID              string  `json:"id"`
		Description     string  `json:"description"`
		LongDescription string  `json:"long_description"`
		Points          float32 `json:"points"`
	} `json:"rubric"`
}

func (c *Client) GetAssignments(courseId int) ([]AssignmentResponse, error) {
	url := config.BaseURL + "/api/v1/courses/" + strconv.Itoa(courseId) + "/assignments?per_page=100&include[]=submission"

	resp, err := c.Fetch(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out []AssignmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, err
}
