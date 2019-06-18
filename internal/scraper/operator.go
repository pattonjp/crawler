package scraper

import (
	"fmt"

	"github.com/pattonjp/crawler/internal"
)

//Operator scrapes pages from in channel and sends result to out channel
func Operator(in chan *internal.ScrapeRequest, out chan *internal.ScrapeResponse, maxDepth int) {
	for req := range in {
		s, err := NewScraper(req.URL)

		if err != nil {
			fmt.Println(err)
		} else {
			resp := &internal.ScrapeResponse{
				Page:    s.GetPage(),
				Content: s.Body(),
			}
			out <- resp
			//Add additional links if not at maxDepth
			if req.Depth < maxDepth {
				// fmt.Printf("Creating %d new requests. depth %d \n", len(s.Links()), req.Depth)
				for _, link := range s.Links() {
					next := &internal.ScrapeRequest{
						URL:   link,
						Depth: req.Depth + 1,
					}
					go func() { in <- next }()
				}
			}
		}
	}
}
