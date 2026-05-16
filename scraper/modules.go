package scraper

import (
	"log"

	"github.com/andrewyur/canvas-scraper-go/pathbuilder"
	"github.com/andrewyur/canvas-scraper-go/writer"
	"golang.org/x/sync/errgroup"
)

func (s *Scraper) scrapeModules(path pathbuilder.PathBuilder, courseId int) error {
	apiResponse, err := s.apiClient.GetModules(courseId)
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, module := range apiResponse {
		modulePath := path.Fork("Modules", module.Name)
		for _, item := range module.Items {
			item := item
			g.Go(func() error {
				if item.Type != "File" {
					return nil
				}
				fileResponse, err := s.apiClient.GetFileDownloadUrl(item.Url)
				if err != nil {
					log.Println("Error getting download Url of file:", item.Url, err)
					return nil
				}
				if fileResponse.Url == nil {
					log.Println("File locked, unable to get download Url:", item.Url, err)
				}
				s.writer.Download(writer.DownloadJob{
					DownloadUrl: *fileResponse.Url,
					Path:        modulePath.Build(),
				})
				return nil
			})
		}
	}
	return g.Wait()
}
