package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// type url string

func fatalIfError(err error, fatal bool) {
	if err != nil {
		fmt.Println(err)
		if fatal == true {
			os.Exit(1)
		}
	}

}

type globalCounter struct {
	mu  sync.Mutex
	val int
	max int
}

func (c *globalCounter) safeInc() {
	c.mu.Lock()
	c.val++
	defer c.mu.Unlock()
}

// var gc globalCounter

func main() {
	gc := globalCounter{
		sync.Mutex{},
		0,
		0,
	}
	var wg sync.WaitGroup
	jobs := make(chan string)
	filepath := flag.String("file", "", "path to a text file containing the urls")
	flag.Parse()
	for w := 1; w <= 3; w++ {
		wg.Go(func() {
			worker(jobs, &gc)
		})
	}

	if *filepath == "" {
		loopOverArgs(jobs, &gc)
	} else {
		loopOverURLs(extractUrls(filepath), jobs, &gc)
	}

	close(jobs)
	wg.Wait()

	if gc.max == gc.val {
		os.Exit(0)
	} else {
		os.Exit(1)
	}

}

func worker(jobs <-chan string, gc *globalCounter) {
	for url := range jobs {
		ping(url, gc)

	}
}

func extractUrls(filepath *string) []string {
	contents, err := os.ReadFile(*filepath)
	fatalIfError(err, true)
	urls := strings.Split(string(contents), "\n")
	return urls
}

func loopOverURLs(urls []string, jobs chan<- string, gc *globalCounter) {
	gc.max = (len(urls))
	for _, url := range urls {
		jobs <- url
	}
}

func loopOverArgs(jobs chan<- string, gc *globalCounter) {
	gc.max = (len(os.Args[1:]))
	for _, url := range os.Args[1:] {
		jobs <- url
	}
}

func ping(url string, gc *globalCounter) {
	start := time.Now()
	client := http.Client{
		Timeout: 9 * time.Second,
	}
	resp, err := client.Get(url)
	end := time.Since(start)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		flush(url, false, "Timeout", end)
		return
	default:
		fatalIfError(err, false)
	}

	// Strict Health Check - 2xx only
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		flush(url, true, resp.Status, end)
		gc.safeInc()
	} else {
		flush(url, false, resp.Status, end)
	}
	defer resp.Body.Close()
}

func flush(url string, ok bool, why string, end time.Duration) {
	var okString string
	switch ok {
	case true:
		okString = "Healthy"
	case false:
		okString = "Unhealthy"
	}
	fmt.Println(url, okString, why, end)
}
