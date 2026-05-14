package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/andrewyur/canvas-scraper-go/config"
	"github.com/andrewyur/canvas-scraper-go/requests"
)

type rawAssignment struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	PointsPossible float32 `json:"points_possible"`
	Rubric         []struct {
		Description     string `json:"description"`
		LongDescription string `json:"long_description"`
	} `json:"rubric"`
	Submission struct {
		Score       float32 `json:"score"`
		Attachments []struct {
			Url string `json:"url"`
		} `json:"attachments"`
	} `json:"submission"`
}

type rubric struct {
	Description     string
	LongDescription string
}
type Assignment struct {
	ID          int
	Name        string
	Description string
	Files       []string
	Score       float32
	Rubric      []rubric
	Attachments []string
}

func GetAssignments(client *http.Client, token, course string) ([]Assignment, error) {
	url := config.BaseURL + "/api/v1/courses/" + course + "/assignments?per_page=100&include[]=submission"

	resp, err := requests.Fetch(client, token, url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var rawAssignments []rawAssignment
	if err := json.NewDecoder(resp.Body).Decode(&rawAssignments); err != nil {
		return nil, err
	}

	assignments := make([]Assignment, 0, len(rawAssignments))
	for _, raw := range rawAssignments {
		assignment, keep := parseAssignment(raw, course)
		if !keep {
			continue
		}

		downloadUrls := make([]string, 0)
		for _, link := range assignment.Files {
			downloadUrl, err := getFileDownloadUrl(client, token, link)
			if err != nil {
				// idk how this should be handled
				log.Println("Couldn't get download url for file:", link, err)
			} else {
				downloadUrls = append(downloadUrls, downloadUrl)
			}
		}
		assignment.Files = downloadUrls

		assignments = append(assignments, assignment)
	}

	return assignments, err
}

func parseAssignment(raw rawAssignment, course string) (Assignment, bool) {
	if len(raw.Submission.Attachments) == 0 {
		return Assignment{}, false
	}

	filesUrl := config.BaseURL + "/api/v1/courses/" + course + "/files"
	pattern := regexp.MustCompile(regexp.QuoteMeta(filesUrl) + `[^"]*`)
	links := pattern.FindAllString(raw.Description, -1)

	rubrics := make([]rubric, 0, len(raw.Rubric))
	for _, r := range raw.Rubric {
		rubrics = append(rubrics, rubric{
			Description:     r.Description,
			LongDescription: r.LongDescription,
		})
	}

	attachments := make([]string, 0, len(raw.Submission.Attachments))
	for _, a := range raw.Submission.Attachments {
		attachments = append(attachments, a.Url)
	}

	return Assignment{
		ID:          raw.ID,
		Name:        raw.Name,
		Description: raw.Description,
		Files:       links,
		Score:       (raw.Submission.Score / raw.PointsPossible) * 100,
		Rubric:      rubrics,
		Attachments: attachments,
	}, true
}
