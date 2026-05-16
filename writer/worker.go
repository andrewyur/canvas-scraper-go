package writer

import (
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andrewyur/canvas-scraper-go/pathbuilder"
)

func (w *Writer) startWorker() error {
	downloads := w.downloadJobs
	writes := w.writeJobs

	for writes != nil || downloads != nil {
		select {
		case downloadJob, ok := <-downloads:
			if !ok {
				downloads = nil
				continue
			}
			if err := w.executeDownloadJob(downloadJob); err != nil {
				log.Println("Error running download job:", err)
			}

		case writeJob, ok := <-writes:
			if !ok {
				writes = nil
				continue
			}
			if err := w.executeWriteJob(writeJob); err != nil {
				log.Println("Error running write job:", err)
			}
		}
	}
	return nil
}

const maxContentSize = 60 * 1024 * 1024 // 60MB

func (w *Writer) executeDownloadJob(job DownloadJob) error {
	if err := os.MkdirAll(job.Path, 0775); err != nil {
		return err
	}
	resp, err := w.apiClient.Fetch(job.DownloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// skips mostly videos
	contentSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if contentSize > maxContentSize {
		if resp.Header.Get("Content-Type") != "video/mp4" {
			log.Println(
				"Max content size exceeded",
				resp.Header.Get("Content-Type"),
				resp.Header.Get("Content-Disposition"),
				contentSize/(1024*1024), "MB",
			)
		}
		// we don't consider this an error
		return nil
	}

	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return err
	}
	filename, present := params["filename"]
	if !present {
		filename = "unnamed-file"
	}

	location := filepath.Join(job.Path, pathbuilder.Sanitize(filename))
	f, err := os.Create(location)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func (w *Writer) executeWriteJob(job WriteJob) error {
	if err := os.MkdirAll(job.Path, 0775); err != nil {
		return err
	}

	location := filepath.Join(job.Path, job.Filename)
	f, err := os.Create(location)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, job.Data)
	return err
}
