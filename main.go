package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"github.com/andrewyur/canvas-scraper-go/api"
)

const baseUrl = "https://sit.instructure.com"

func main() {
	client := &http.Client{}

	var (
		token           string
		selectedCourses []int
		courses         []api.Course
		folderConfirm   bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter 'canvas_session' value:").
				Value(&token).
				Description("Download a cookie inspector extension, and look for a cookie named 'canvas_session' when you are on the canvas page. Copy the value, and paste it here.").
				Placeholder("ng9sjIp1SCIBuZzZwVL7Bw+lQPddj0IaBTBmNp2U1KyHobDse2EkTlRS49G9YKuuwTxQF5E-MjM9VT...").
				Validate(func(str string) error {
					strParts := strings.Split(str, ".")
					if len(strParts) != 3 {
						return errors.New("Incomplete token - make sure you have copied the entire thing!")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Which courses do you want to scrape?").
				Value(&selectedCourses).
				Height(30).
				OptionsFunc(func() []huh.Option[int] {
					var err error
					courses, err = api.GetCourses(client, token)

					if err != nil {
						log.Fatal("Could not get courses:", err)
					}

					options := make([]huh.Option[int], 0, len(courses))
					slices.Reverse(courses)
					for i, course := range courses {
						options = append(options, huh.NewOption(course.Name, i).Selected(!course.Concluded))
					}
					return options
				}, &token),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Is it ok that courses are saved to a folder named 'courses' in the current directory?").
				Value(&folderConfirm),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	if !folderConfirm {
		fmt.Println("Did not scrape courses. Exiting...")
		os.Exit(0)
	}

	start := time.Now()

	err := spinner.New().
		Type(spinner.Line).
		Title(" Scraping Courses...").
		Action(func() {
			var wg sync.WaitGroup

			user, err := api.GetUser(client, token)
			if err != nil {
				log.Fatal("Could not get user", err)
			}

			for _, i := range selectedCourses {
				wg.Add(1)
				go func() {
					defer wg.Done()
					scrapeCourse(client, courses[i], user, token, "courses", &wg)
				}()
			}

			wg.Wait()
		}).
		Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Done in %.2fs. upload the courses directory into the Upload folder in the Coursework drive, and run the script\n", time.Since(start).Seconds())
}

var sanitizeRegexp = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func sanitize(s string) string {
	safe := sanitizeRegexp.ReplaceAllString(s, "_")
	safe = strings.Trim(safe, ".")
	return safe
}

func scrapeCourse(client *http.Client, course api.Course, user api.User, token, basePath string, wg *sync.WaitGroup) {
	pathParts := []string{
		basePath,
		course.Department,
		course.Name,
		course.Semester,
		course.Teacher,
		user.Name + " (" + course.Grade + ")",
	}

	for i, part := range pathParts {
		pathParts[i] = sanitize(part)
	}

	coursePath := filepath.Join(pathParts...)

	wg.Add(1)
	go func() {
		defer wg.Done()
		assignments, err := api.GetAssignments(client, token, course.ID)
		if err != nil {
			log.Println("Error fetching assignments:", err)
		}

		for _, assignment := range assignments {
			dirName := sanitize(assignment.Name) + " (" + fmt.Sprintf("%.2f%%", assignment.Score) + ")"
			assignmentPath := filepath.Join(coursePath, "Assignments", dirName)
			// create all directories at once
			err := os.MkdirAll(filepath.Join(assignmentPath, "Submission"), 0755)
			if err != nil {
				log.Println("Could not create assignment folder:", err)
			} else {
				scrapeAssignment(client, assignment, token, assignmentPath, wg)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		modules, err := api.GetModules(client, token, course.ID)
		if err != nil {
			log.Println("Error fetching modules:", err)
		}

		for _, module := range modules {
			modulePath := filepath.Join(coursePath, "Modules", sanitize(module.Name))
			err := os.MkdirAll(filepath.Join(modulePath), 0755)
			if err != nil {
				log.Println("Could not create Module folder:", err)
			} else {
				for _, item := range module.Files {
					wg.Add(1)
					go func() {
						defer wg.Done()

						resp, err := FetchFile(client, item, token)
						if err != nil {
							log.Println("Error fetching file:", err)
							return
						}

						if err := saveResponse(resp, modulePath); err != nil {
							log.Println("Error saving response to file:", err)
							return
						}
					}()
				}
			}
		}
	}()
}

func scrapeAssignment(client *http.Client, assignment api.Assignment, token, path string, wg *sync.WaitGroup) {
	// files
	for _, url := range assignment.Files {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := FetchFile(client, url, token)
			if err != nil {
				log.Println("Error fetching file:", err)
				return
			}

			if err := saveResponse(resp, path); err != nil {
				log.Println("Error saving response to file:", err)
				return
			}
		}()
	}

	//attachments
	for _, url := range assignment.Attachments {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := FetchFile(client, url, token)
			if err != nil {
				log.Println("Error fetching file:", err)
				return
			}

			if err := saveResponse(resp, filepath.Join(path, "Submission")); err != nil {
				log.Println("Error saving file:", err)
				return
			}
		}()
	}

	// description
	f, err := os.Create(filepath.Join(path, "description.html"))
	if err != nil {
		log.Println("Error creating description file:", err)
	} else {
		defer f.Close()
		if _, err := io.WriteString(f, assignment.Description); err != nil {
			log.Println("Error writing to description file:", err)
		}
	}

	// rubric
	if len(assignment.Rubric) > 0 {
		f, err := os.Create(filepath.Join(path, "rubric.txt"))
		if err != nil {
			log.Println("Error writing rubric file:", err)
		} else {
			defer f.Close()
			for _, r := range assignment.Rubric {
				if _, err := io.WriteString(f, r.Description+"\n - "+r.LongDescription+"\n"); err != nil {
					log.Println("Error writing to rubric file:", err)
				}
			}
		}
	}
}

func FetchFile(client *http.Client, url, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "canvas_session",
		Value: token,
	})

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
	}

	return resp, nil
}

func saveResponse(resp *http.Response, path string) error {
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return err
	}
	filename := params["filename"]

	location := filepath.Join(path, sanitize(filename))
	f, err := os.Create(location)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
