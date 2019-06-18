package indexer

import (
	"fmt"
	"regexp"

	"github.com/pattonjp/crawler/internal"
)

//WordIdx word index of all scrapped pages
type WordIdx struct {
	Page  *internal.Page `json:"page"`
	Count int            `json:"count"`
}

//Index holds
type Index struct {
	items     map[string]map[string]*WordIdx
	channel   chan *internal.ScrapeResponse
	pages     map[string]bool
	stopWords map[string]struct{}
}

//Meta data structure for returning index info
type Meta struct {
	Words int `json:"words"`
	Pages int `json:"pages"`
}

func (idx *WordIdx) String() string {
	p := idx.Page
	return fmt.Sprintf("%s \n\t Count: %d\n\t %s \n\t %s", p.URL, idx.Count, p.Title, p.Description)
}

//NewIndexer creates a new indexer
func NewIndexer(in chan *internal.ScrapeResponse, stopWords []string) *Index {
	m := make(map[string]struct{}, len(stopWords))
	for _, sw := range stopWords {
		m[sw] = struct{}{}
	}
	return &Index{
		items:     make(map[string]map[string]*WordIdx),
		channel:   in,
		pages:     make(map[string]bool),
		stopWords: m,
	}
}

//HandleAdditions indexes a page when added to channel
func (idx *Index) HandleAdditions() {
	for page := range idx.channel {
		idx.add(page.Page, page.Content)
	}
}

func (idx *Index) add(page *internal.Page, content string) {
	wordsExp := regexp.MustCompile(`\w+`)
	words := wordsExp.FindAllString(content, -1)
	wm := make(map[string]int)
	for _, word := range words {
		wm[word]++
	}
	for key, val := range wm {
		if _, ok := idx.stopWords[key]; ok {
			continue
		}
		_, ok := idx.items[key]
		if !ok {
			idx.items[key] = make(map[string]*WordIdx)
		}
		idx.items[key][page.URL] = &WordIdx{page, val}
	}
	idx.pages[page.URL] = true
}

//Search gets pages by term from items map
func (idx *Index) Search(str string) []*WordIdx {

	var res []*WordIdx
	for _, val := range idx.items[str] {
		res = append(res, val)
	}
	return res
}

//Close dispose of any server resources
func (idx *Index) Close() {

}

//Reset resets index
func (idx *Index) Reset() {
	idx.pages = nil
	idx.items = nil
	idx.items = make(map[string]map[string]*WordIdx)
	idx.pages = make(map[string]bool)
}

//Meta meta information about the index
func (idx *Index) Meta() *Meta {
	return &Meta{
		Words: len(idx.items),
		Pages: len(idx.pages),
	}

}
