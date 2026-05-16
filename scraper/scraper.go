package scraper

import (
	"github.com/andrewyur/canvas-scraper-go/api"
	"github.com/andrewyur/canvas-scraper-go/writer"
)

type Scraper struct {
	apiClient *api.Client
	writer    *writer.Writer
	userName  string
}

func NewScraper(apiClient *api.Client, maxContentSize uint64) (*Scraper, error) {
	userResponse, err := apiClient.GetUser()
	if err != nil {
		return nil, err
	}

	return &Scraper{
		apiClient: apiClient,
		userName:  userResponse.Name,
		writer:    writer.CreateWriter(apiClient, maxContentSize),
	}, nil
}

func (s *Scraper) Close() {
	s.writer.Close()
}
