package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type unparsedCourse struct {
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

type Course struct {
	ID         string
	Department string
	Name       string
	Concluded  bool
	Grade      string
	Teacher    string
	Semester   string
}

func GetCourses(client *http.Client, token string) ([]Course, error) {
	url := config.BaseURL + "/api/v1/courses?include[]=teachers&include[]=total_scores&include[]=concluded&include[]=term&per_page=100"

	body, err := fetchJSON(client, token, url)
	if err != nil {
		return nil, err
	}

	defer body.Close()

	var rawCourses []unparsedCourse
	if err := json.NewDecoder(body).Decode(&rawCourses); err != nil {
		return nil, err
	}

	courses := make([]Course, 0, len(rawCourses))
	for _, raw := range rawCourses {
		if course, returned := parseCourse(raw); returned {
			courses = append(courses, course)
		}
	}

	return courses, nil
}

// expecting course name to be in format 20XXS CS 115
func parseCourse(raw unparsedCourse) (Course, bool) {
	if raw.AccessRestrictedByDate != nil || len(raw.Teachers) == 0 || raw.Term.ID == 1 {
		return Course{}, false
	}

	nameParts := strings.Split(raw.Name, " ")
	var semester string
	if nameParts[0][4] == 'F' {
		semester = nameParts[0][:4] + " Fall"
	} else {
		semester = nameParts[0][:4] + " Spring"
	}

	numberParts := strings.Split(nameParts[2], "-")
	name := nameParts[1] + " " + numberParts[0] + " - " + raw.CourseCode

	var grade string
	if len(raw.Enrollments) > 0 && raw.Enrollments[0].ComputedCurrentScore != nil {
		grade = fmt.Sprintf("%.2f%%", *raw.Enrollments[0].ComputedCurrentScore)
	} else {
		grade = "N/A"
	}

	// sometimes teachers can have their names revoked, and their last name is a '-' instead
	teachers := make([]string, 0)
	for _, t := range raw.Teachers {
		nameParts := strings.Split(t.DisplayName, " ")
		lastName := nameParts[len(nameParts)-1]
		if strings.Trim(lastName, " ") != "-" {
			teachers = append(teachers, nameParts[len(nameParts)-1])
		}
	}

	return Course{
		ID:         strconv.Itoa(raw.ID),
		Department: nameParts[1],
		Name:       name,
		Concluded:  raw.Concluded,
		Grade:      grade,
		Teacher:    strings.Join(teachers, " - "),
		Semester:   semester,
	}, true
}
