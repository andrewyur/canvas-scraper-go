package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type rawModule struct {
	Name  string `json:"name"`
	Items []struct {
		Url  string `json:"url"`
		Type string `json:"type"`
	} `json:"items"`
}

type Module struct {
	Name  string
	Files []string
}

func GetModules(client *http.Client, token, course string) ([]Module, error) {
	url := config.BaseURL + "/api/v1/courses/" + course + "/modules?include[]=items&include[]=content_details&per_page=30"

	body, err := fetchJSON(client, token, url)
	if err != nil {
		return nil, err
	}

	defer body.Close()

	var rawModules []rawModule
	if err := json.NewDecoder(body).Decode(&rawModules); err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(rawModules))
	for _, raw := range rawModules {
		files := make([]string, 0, len(raw.Items))
		for _, item := range raw.Items {
			if item.Type != "File" {
				continue
			}

			if downloadUrl, err := getFileDownloadUrl(client, token, item.Url); err != nil {
				log.Println("Couldn't get file download url:", downloadUrl)
			} else {
				files = append(files, downloadUrl)
			}
		}

		if len(files) > 0 {
			modules = append(modules, Module{
				Name:  raw.Name,
				Files: files,
			})
		}
	}

	return modules, nil
}
