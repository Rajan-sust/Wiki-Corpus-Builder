package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"math"
)

type WikiResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

type RateLimiter struct {
	maxRequests    int
	currentRequests int
	startTime      time.Time
}

func NewRateLimiter(maxRequests int) *RateLimiter {
	return &RateLimiter{
		maxRequests:    maxRequests,
		currentRequests: 0,
		startTime:      time.Now(),
	}
}

func (rl *RateLimiter) Wait() {
	elapsed := time.Since(rl.startTime)

	if rl.currentRequests >= rl.maxRequests {
		// sleepTime := time.Hour - elapsed
		sleepTime := time.Duration(math.Max(0, float64(time.Hour - elapsed)))
		time.Sleep(sleepTime)
		// Reset after waiting
		rl.currentRequests = 0
		rl.startTime = time.Now()
	}

	rl.currentRequests++
}

func main() {
	lang := flag.String("lang", "", "Language code (e.g., bn)")
	inputFile := flag.String("input", "", "Input file with titles")
	outputFile := flag.String("output", "", "Output file to save extracts")
	flag.Parse()

	if *lang == "" || *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: go run main.go --lang=bn --input titles.txt --output wiki.txt")
		os.Exit(1)
	}

	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	inputHandle, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer inputHandle.Close()

	outputHandle, err := os.OpenFile(*outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening output file: %v\n", err)
		os.Exit(1)
	}
	defer outputHandle.Close()

	rateLimiter := NewRateLimiter(4990)

	scanner := bufio.NewScanner(inputHandle)
	for scanner.Scan() {
		title := strings.TrimSpace(scanner.Text())
		if title == "" {
			continue
		}
		
		rateLimiter.Wait()

		extract, err := fetchWikipediaExtract(*lang, title)
		if err != nil {
			fmt.Printf("Error fetching extract for %s: %v\n", title, err)
			continue
		}

		_, err = outputHandle.WriteString(fmt.Sprintf("%s\n", extract))
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
		} else {
			fmt.Printf("Page `%s` successfully fetched\n", title)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Wikipedia extracts saved successfully.")
}

func fetchWikipediaExtract(lang, title string) (string, error) {
	encodedTitle := url.QueryEscape(title)
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exintro&explaintext&redirects=1&titles=%s", lang, encodedTitle)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var wikiResp WikiResponse
	if err := json.Unmarshal(body, &wikiResp); err != nil {
		return "", err
	}

	var extract string
	for _, page := range wikiResp.Query.Pages {
		extract = strings.ReplaceAll(page.Extract, "\n", " ")
		break
	}

	if extract == "" {
		return "", fmt.Errorf("no extract found for title: %s", title)
	}

	return extract, nil
}