package api

import (
	"encoding/json"
	"strconv"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type SubmissionResponse struct {
	Score          float32 `json:"score"`
	SubmissionType string  `json:"submission_type"`
	Comments       []struct {
		AuthorName  string `json:"author_name"`
		Comment     string `json:"comment"`
		Attachments []struct {
			Url string `json:"url"`
		} `json:"attachments"`
	} `json:"submission_comments"`
	Attachments []struct {
		Url string `json:"url"`
	} `json:"attachments"`
	RubricAssessment map[string]struct {
		Comments string  `json:"comments"`
		Points   float32 `json:"points"`
	} `json:"rubric_assessment"`
}

func (c *Client) GetSubmission(courseId, assignmentId int) (SubmissionResponse, error) {
	url := config.BaseURL + "/api/v1/courses/" + strconv.Itoa(courseId) + "/assignments/" + strconv.Itoa(assignmentId) + "/submissions/self?include[]=rubric_assessment&include[]=submission_comments"

	resp, err := c.Fetch(url)
	if err != nil {
		return SubmissionResponse{}, err
	}

	defer resp.Body.Close()

	var out SubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return SubmissionResponse{}, err
	}

	return out, nil
}
