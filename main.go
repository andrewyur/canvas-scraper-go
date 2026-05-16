package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"github.com/andrewyur/canvas-scraper-go/api"
	"github.com/andrewyur/canvas-scraper-go/pathbuilder"
	"github.com/andrewyur/canvas-scraper-go/scraper"
)

const defaultLogFilePath = ".canvas-log.txt"

func main() {
	httpClient := &http.Client{Timeout: 45 * time.Second}

	var logFilePath string
	if value, set := os.LookupEnv("CANVAS_LOGFILE"); set {
		logFilePath = value
	} else {
		logFilePath = defaultLogFilePath
	}

	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatal("Could not open log file: ", err)
	}
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Fatal("Could not create log file: ", err)
	}
	log.SetOutput(logFile)
	defer logFile.Close()

	var (
		token           string
		selectedCourses []int
		courses         []scraper.CourseInfo
		folderConfirm   bool
		scraperHandle   *scraper.Scraper
		apiClient       *api.Client
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
				Description("Scroll with the arrow keys to access all courses, some may be obscured").
				Value(&selectedCourses).
				Height(30).
				OptionsFunc(func() []huh.Option[int] {
					var err error
					apiClient = api.NewClient(httpClient, token)
					scraperHandle, err = scraper.NewScraper(apiClient)

					if err != nil {
						log.Fatal("Could not create scraper:", err)
					}

					courses, err = scraperHandle.GetCourseInfo()

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

	err = spinner.New().
		Type(spinner.Line).
		Title(" Scraping Courses...").
		ActionWithErr(func(_ context.Context) error {

			var selectedCourseInfo []scraper.CourseInfo

			for _, i := range selectedCourses {
				selectedCourseInfo = append(selectedCourseInfo, courses[i])
			}

			basePath := pathbuilder.NewPathBuilder("courses")

			err := scraperHandle.ScrapeCourses(basePath, selectedCourseInfo)

			scraperHandle.Close()

			return err
		}).
		Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scraped %d courses in %.2fs. upload the courses directory into the Upload folder in the Coursework drive, and run the script\n", len(selectedCourses), time.Since(start).Seconds())
}
