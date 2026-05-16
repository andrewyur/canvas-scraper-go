package scraper

import (
	"fmt"
	"strings"

	"github.com/andrewyur/canvas-scraper-go/api"
)

type CourseInfo struct {
	ID         int
	Department string
	Name       string
	Concluded  bool
	Grade      string
	Teacher    string
	Semester   string
}

func (s *Scraper) GetCourseInfo() ([]CourseInfo, error) {
	apiResponse, err := s.apiClient.GetCourses()
	if err != nil {
		return nil, err
	}

	info := make([]CourseInfo, 0)
	for _, raw := range apiResponse {
		if course, keep := parseCourse(raw); keep {
			info = append(info, course)
		}
	}

	return info, nil
}

// expecting course name to be in format 20XXS CS 115
func parseCourse(raw api.CourseResponse) (CourseInfo, bool) {
	if raw.AccessRestrictedByDate != nil || len(raw.Teachers) == 0 || raw.Term.ID == 1 {
		return CourseInfo{}, false
	}

	nameParts := strings.Split(raw.Name, " ")

	if len(nameParts) != 3 || len(nameParts[0]) != 5 {
		return CourseInfo{}, false
	}

	var semester string
	if nameParts[0][4] == 'F' {
		semester = nameParts[0][:4] + " Fall"
	} else {
		semester = nameParts[0][:4] + " Spring"
	}

	numberParts := strings.Split(nameParts[2], "-")
	name := nameParts[1] + " " + numberParts[0] + " - " + raw.CourseCode

	// Sometimes teachers mess up the naming conventions
	department := strings.Split(nameParts[1], "-")[0]

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

	return CourseInfo{
		ID:         raw.ID,
		Department: department,
		Name:       name,
		Concluded:  raw.Concluded,
		Grade:      grade,
		Teacher:    strings.Join(teachers, " - "),
		Semester:   semester,
	}, true
}
