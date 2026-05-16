package scraper

import (
	"github.com/andrewyur/canvas-scraper-go/pathbuilder"
	"golang.org/x/sync/errgroup"
)

func (s *Scraper) ScrapeCourses(path pathbuilder.PathBuilder, courses []CourseInfo, scrapeModules bool) error {
	g := errgroup.Group{}

	for _, course := range courses {
		courseId := course.ID
		coursePath := path.Fork(
			course.Department,
			course.Name,
			course.Semester,
			course.Teacher,
			s.userName+" ("+course.Grade+")",
		)
		g.Go(func() error {
			return s.scrapeAssignments(coursePath, courseId)
		})

		if scrapeModules {
			g.Go(func() error {
				return s.scrapeModules(coursePath, courseId)
			})
		}
	}
	return g.Wait()
}
