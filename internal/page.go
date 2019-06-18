package internal

//Page represents a scrapped page
type Page struct {
	URL         string `json:"page"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

//ScrapeRequest data for scraping url
type ScrapeRequest struct {
	URL   string
	Depth int
}

//ScrapeResponse results of a page scrape
type ScrapeResponse struct {
	Page    *Page
	Content string
}
