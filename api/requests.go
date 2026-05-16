package api

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func (c *Client) Fetch(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "canvas_session",
		Value: c.token,
	})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		resp.Body.Close()

		wait := time.Second + time.Duration(rand.Int()%3000)*time.Millisecond

		log.Println("Rate limited, waiting", wait)
		time.Sleep(wait)
		return c.Fetch(url)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
	}

	return resp, nil
}

func (c *Client) FetchText(url string) (string, error) {
	resp, err := c.Fetch(url)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), nil
}
