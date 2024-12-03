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
	"strings"
)

type WikiResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

func main() {
	// Define command-line flags
	lang := flag.String("lang", "", "Language code (e.g., bn)")
	inputFile := flag.String("input", "", "Input file with titles")
	outputFile := flag.String("output", "", "Output file to save extracts")
	flag.Parse()

	// Validate required flags
	if *lang == "" || *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: go run main.go --lang=bn --input titles.txt --output wiki.txt")
		os.Exit(1)
	}

	// Open input file
	inputHandle, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer inputHandle.Close()

	// Open output file
	outputHandle, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputHandle.Close()

	// Read titles from input file
	scanner := bufio.NewScanner(inputHandle)
	for scanner.Scan() {
		title := strings.TrimSpace(scanner.Text())
		if title == "" {
			continue
		}

		// Fetch extract for the title
		extract, err := fetchWikipediaExtract(*lang, title)
		if err != nil {
			fmt.Printf("Error fetching extract for %s: %v\n", title, err)
			continue
		}

		// Write to output file
		_, err = outputHandle.WriteString(fmt.Sprintf("%s:\n%s\n\n", title, extract))
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Wikipedia extracts saved successfully.")
}

func fetchWikipediaExtract(lang, title string) (string, error) {
	// Construct API URL
	encodedTitle := url.QueryEscape(title)
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exintro&explaintext&redirects=1&titles=%s", lang, encodedTitle)

	// Send HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse JSON response
	var wikiResp WikiResponse
	err = json.Unmarshal(body, &wikiResp)
	if err != nil {
		return "", err
	}

	// Extract first (and only) page from the response
	for _, page := range wikiResp.Query.Pages {
		// Remove newlines and return cleaned extract
		return strings.ReplaceAll(page.Extract, "\n", " "), nil
	}

	return "", fmt.Errorf("no extract found for title: %s", title)
}