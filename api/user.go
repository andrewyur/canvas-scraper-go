package api

import (
	"encoding/json"
	"net/http"

	"github.com/andrewyur/canvas-scraper-go/config"
	"github.com/andrewyur/canvas-scraper-go/requests"
)

type User struct {
	Name string `json:"short_name"`
}

func GetUser(client *http.Client, token string) (User, error) {
	url := config.BaseURL + "/api/v1/users/self"

	resp, err := requests.Fetch(client, token, url)
	if err != nil {
		return User{}, err
	}

	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}
