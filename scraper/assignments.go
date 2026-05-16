package scraper

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/andrewyur/canvas-scraper-go/api"
	"github.com/andrewyur/canvas-scraper-go/config"
	"github.com/andrewyur/canvas-scraper-go/pathbuilder"
	"github.com/andrewyur/canvas-scraper-go/writer"
	"golang.org/x/sync/errgroup"
)

func (s *Scraper) scrapeAssignments(path pathbuilder.PathBuilder, courseId int) error {
	assignments, err := s.apiClient.GetAssignments(courseId)
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, assignment := range assignments {
		if !slices.Contains(assignment.SubmissionTypes, "online_upload") {
			continue
		}

		assignment := assignment
		g.Go(func() error {
			submission, err := s.apiClient.GetSubmission(courseId, assignment.ID)

			if err != nil {
				log.Println("Could not get submission for assignment:", err)
				return nil
			}

			grade := (submission.Score / assignment.PointsPossible) * 100
			var assignmentFolderName string
			if math.IsInf(float64(grade), 0) {
				assignmentFolderName = fmt.Sprintf("%s (NA)", assignment.Name)
			} else {
				assignmentFolderName = fmt.Sprintf("%s (%.0f%%)", assignment.Name, grade)
			}

			assignmentPath := path.Fork("Assignments", assignmentFolderName)

			// don't make network requests, so don't need a goroutine
			s.scrapeAssignmentRubric(assignmentPath, assignment, submission)
			s.scrapeAssignmentSubmission(assignmentPath, submission)

			var subg errgroup.Group
			subg.Go(func() error {
				s.scrapeAssignmentComments(assignmentPath, submission)
				return nil
			})
			subg.Go(func() error {
				s.scrapeAssignmentDescription(assignmentPath, assignment, courseId)
				return nil
			})

			return subg.Wait()
		})
	}
	return g.Wait()
}

func (s *Scraper) scrapeAssignmentComments(path pathbuilder.PathBuilder, submission api.SubmissionResponse) {
	var g errgroup.Group
	comments := make([]string, len(submission.Comments))
	for i, comment := range submission.Comments {
		i, comment := i, comment
		g.Go(func() error {
			text := fmt.Sprintf("%s: %s\n", comment.AuthorName, comment.Comment)
			for _, attachment := range comment.Attachments {
				// comment attachments are pretty short, so we circumvent the writer/downloader for this
				attachmentContent, err := s.apiClient.FetchText(attachment.Url)
				if err != nil {
					log.Println("Could not get attachment content", attachment.Url, err)
					continue
				}
				text += fmt.Sprintf("ATTACHMENT:\n%s\n", attachmentContent)
			}
			comments[i] = text
			return nil
		})
	}
	g.Wait()
	s.writer.Write(writer.WriteJob{
		Path:     path.Build(),
		Filename: "comments.txt",
		Data:     strings.Join(comments, "\n"),
	})
}

func (s *Scraper) scrapeAssignmentRubric(path pathbuilder.PathBuilder, assignment api.AssignmentResponse, submission api.SubmissionResponse) {
	var assessments []string
	for _, criterion := range assignment.Rubric {
		grade := submission.RubricAssessment[criterion.ID]
		text := fmt.Sprintf("CRITERIA: %s - %s\n", criterion.Description, criterion.LongDescription)
		text += fmt.Sprintf("GRADE: %.1f/%.1f\n", grade.Points, criterion.Points)
		text += fmt.Sprintf("COMMENTS:\n%s\n", grade.Comments)
		assessments = append(assessments, text)
	}
	s.writer.Write(writer.WriteJob{
		Path:     path.Build(),
		Filename: "rubric.txt",
		Data:     strings.Join(assessments, "\n"),
	})
}

func (s *Scraper) scrapeAssignmentDescription(path pathbuilder.PathBuilder, assignment api.AssignmentResponse, courseId int) {
	// looking specifically for file api links here
	filesUrl := config.BaseURL + "/api/v1/courses/" + strconv.Itoa(courseId) + "/files"
	pattern := regexp.MustCompile(regexp.QuoteMeta(filesUrl) + `[^"]*`)
	links := pattern.FindAllString(assignment.Description, -1)

	pathString := path.Build()

	var g errgroup.Group
	for _, link := range links {
		link := link
		g.Go(func() error {
			file, err := s.apiClient.GetFileDownloadUrl(link)
			if err != nil || file.Url == nil {
				log.Println("Couldn't get download url for file:", link, err)
				return nil
			}

			s.writer.Download(writer.DownloadJob{
				DownloadUrl: *file.Url,
				Path:        pathString,
			})
			return nil
		})
	}
	s.writer.Write(writer.WriteJob{
		Path:     pathString,
		Filename: "description.html",
		Data:     assignment.Description,
	})
	g.Wait()
}

func (s *Scraper) scrapeAssignmentSubmission(path pathbuilder.PathBuilder, submission api.SubmissionResponse) {
	submissionsPath := path.Fork("Submission").Build()
	for _, attachment := range submission.Attachments {
		s.writer.Download(writer.DownloadJob{
			DownloadUrl: attachment.Url,
			Path:        submissionsPath,
		})
	}
}
