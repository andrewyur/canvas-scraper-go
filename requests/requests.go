package requests

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func doRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		resp.Body.Close()

		wait := time.Second + time.Duration(rand.Int()%500)*time.Millisecond

		log.Println("Rate limited, waiting", wait)
		time.Sleep(wait)
		return doRequest(client, req)
	}

	return resp, nil
}

func Fetch(client *http.Client, token, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "canvas_session",
		Value: token,
	})

	resp, err := doRequest(client, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
	}

	return resp, nil
}
