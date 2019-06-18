package scraper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pattonjp/crawler/internal"
)

// Scraper for each website
type Scraper struct {
	url string
	doc *goquery.Document
}

// NewScraper builds a new scraper for the website
func NewScraper(u string) (*Scraper, error) {
	if !strings.HasPrefix(u, "http") {
		return nil, fmt.Errorf("url %s did not have a scheme", u)
	}

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	d, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	return &Scraper{
		url: u,
		doc: d,
	}, nil
}

// Body returns a string with the body of the page
func (s *Scraper) Body() string {
	body := s.doc.Find("body").Text()
	// Remove leading/ending white spaces
	body = strings.TrimSpace(body)

	return body
}

func (s *Scraper) buildLink(href string) string {
	var link string

	if strings.HasPrefix(href, "/") {
		link = strings.Join([]string{s.url, href}, "")
	} else {
		link = href
	}

	link = strings.TrimRight(link, "/")
	link = strings.TrimRight(link, ":")

	return link
}

// Links returns an array with all the links from the website
func (s *Scraper) Links() []string {
	links := make([]string, 0)
	var link string

	s.doc.Find("body a").Each(func(index int, item *goquery.Selection) {
		link = ""

		linkTag := item
		href, _ := linkTag.Attr("href")

		if !strings.Contains(href, "#") && !strings.HasPrefix(href, "/#") && !strings.HasPrefix(href, "javascript") {
			link = s.buildLink(href)
			if link != "" &&
				!strings.HasPrefix(link, "mailto:") &&
				!strings.HasPrefix(link, "tel:") {
				links = append(links, link)
			}
		}
	})

	return links
}

//Title finds the title of the page
func (s *Scraper) Title() string {
	return s.doc.Find("title").Contents().Text()
}

// Description finds the description of the page
func (s *Scraper) Description() string {
	var d string
	s.doc.Find("meta").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("name", "") == "description" || item.AttrOr("property", "") == "og:description" {
			d = item.AttrOr("content", "")
		}
	})

	return d
}

func (s *Scraper) GetPage() *internal.Page {
	return &internal.Page{
		URL:         s.url,
		Title:       s.Title(),
		Description: s.Description(),
	}
}
