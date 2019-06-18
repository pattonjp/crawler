package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/pattonjp/crawler/internal"
	"github.com/pattonjp/crawler/internal/indexer"
	"github.com/pattonjp/crawler/internal/scraper"
	"github.com/pattonjp/crawler/internal/web"
)

const stopWordPath = "./config/stopwords.ini"

func main() {
	port := flag.String("port", "5000", "port to run the web site on")
	debug := flag.Bool("debug", false, "sets app in debug mode")
	scrappers := flag.Int("concurrency", 5, "number of concurent scrappers to run ")
	flag.Parse()

	reqChan := make(chan *internal.ScrapeRequest)
	respChan := make(chan *internal.ScrapeResponse)

	stopWords, err := getStopWords(stopWordPath)

	if err != nil {
		fmt.Println(err)
		return
	}
	idx := indexer.NewIndexer(respChan, stopWords)
	go idx.HandleAdditions()

	for i := 0; i < *scrappers; i++ {
		go scraper.Operator(reqChan, respChan, 3)
	}

	dir := "./internal/web/templates/"
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	server, err := web.NewServer(idx, reqChan, dir, *debug)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		sig := <-gracefulStop
		server.Close()
		idx.Close()
		fmt.Printf("caught sig: %+v", sig)
		os.Exit(0)
	}()
	err = server.ServeHTTP(":" + *port)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getStopWords(filePath string) ([]string, error) {
	stopWords := []string{}
	f, err := os.Open(filePath)
	if err != nil {
		return stopWords, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		stopWords = append(stopWords, s.Text())
	}

	return stopWords, s.Err()
}
