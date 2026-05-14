package api

import (
	"encoding/json"
	"net/http"

	"github.com/andrewyur/canvas-scraper-go/config"
)

type User struct {
	Name string `json:"short_name"`
}

func GetUser(client *http.Client, token string) (User, error) {
	url := config.BaseURL + "/api/v1/users/self"

	body, err := fetchJSON(client, token, url)
	if err != nil {
		return User{}, err
	}

	defer body.Close()

	var user User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}
