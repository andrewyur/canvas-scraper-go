package writer

import (
	"github.com/andrewyur/canvas-scraper-go/api"
	"golang.org/x/sync/errgroup"
)

type Writer struct {
	apiClient    *api.Client
	downloadJobs chan DownloadJob
	writeJobs    chan WriteJob
	group        *errgroup.Group
}

// relatively arbitrary
const numWorkers = 60

func CreateWriter(apiClient *api.Client) *Writer {
	var g errgroup.Group

	w := &Writer{
		apiClient:    apiClient,
		group:        &g,
		downloadJobs: make(chan DownloadJob),
		writeJobs:    make(chan WriteJob),
	}

	for range numWorkers {
		g.Go(w.startWorker)
	}

	return w
}

func (w *Writer) Close() error {
	close(w.downloadJobs)
	close(w.writeJobs)
	return w.group.Wait()
}

type DownloadJob struct {
	Path        string
	DownloadUrl string
}

func (w *Writer) Download(job DownloadJob) {
	w.downloadJobs <- job
}

type WriteJob struct {
	Path     string
	Filename string
	Data     string
}

func (w *Writer) Write(job WriteJob) {
	w.writeJobs <- job
}
