package main

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

func getListOfEscorts() []string {
	res := []string{}

	for i := 0; i <= 830; i++ {
		str := "https://tryst.link/search?page=" + strconv.Itoa(i)
		res = append(res, str)
	}
	return res
}

func extractEscortName(s string, regexForName regexp.Regexp) string {
	cityName := regexForName.FindString(s)
	if cityName != "" {
		chopOffFront := cityName[7:]
		return chopOffFront
	}
	return ""
}

func isImageWeWant(s string, regexForImage regexp.Regexp) bool {
	image := regexForImage.FindString(s)
	if image != "" {
		return true
	}
	return false
}

//tryst pic scraper
//and rates
//WARNING - terrible code, do not put everything in main
func main() {
	SLEEP_TIME_MS := 1000
	//steps:
	//set up names hashtable
	//generate list of pages to visit (, 0, 2, 830 (https://tryst.link/search?page=0))
	//for each page:
	//  grab all the links and check if we've seen them before (https://tryst.link/escort/lisalangsd)
	//  (cut it down to lisalangsd)
	//  for each escort:
	//     if we havent seen it, visit it
	//     grab https://media.tryst.a4cdn.ch/*********.jpg , also save the text (rates? lol)
	//      ???
	// Things still to do: move sleeptime to constant, and test how much ratelimiting is needed

	regexEtractEscortName := regexp.MustCompile(`escort\/([A-Za-z0-9-]+)$`)
	regexForImage := regexp.MustCompile(`\/media\.tryst\.a4cdn\.ch\/.+\.jpg`)
	listOfPages := getListOfEscorts()
	setOfEscorts := make(map[string]bool)
	currentEscort := ""

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	c := colly.NewCollector(
		colly.MaxDepth(2),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		isImageWeWantNow := isImageWeWant(link, *regexForImage)
		name := extractEscortName(link, *regexEtractEscortName)

		//first check for images
		if isImageWeWantNow {
			//get the image
			fileName := currentEscort + "/" + link[len(link)-15:]
			os.Mkdir(currentEscort, 777)

			newFile, _ := os.Create(fileName)
			resp, _ := client.Get(link)
			_, _ = io.Copy(newFile, resp.Body)
			resp.Body.Close()
			newFile.Close()
			return
		}

		//no escort found
		if name == "" {
			return
		}

		// if weve already seen them, also return
		if _, ok := setOfEscorts[name]; ok {
			return
		}

		setOfEscorts[name] = true
		currentEscort = name
		//make a folder for their name
		time.Sleep(time.Duration(SLEEP_TIME_MS) * time.Millisecond)
		e.Request.Visit(link)
	})

	for _, k := range listOfPages {
		c.Visit(k)
		//dont get 1015 rate limited
		time.Sleep(time.Duration(SLEEP_TIME_MS) * time.Millisecond)
	}

}
