package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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

type LoginTokenResponse struct {
	Query struct {
		Tokens struct {
			LoginToken string `json:"logintoken"`
		} `json:"tokens"`
	} `json:"query"`
}

type LoginResponse struct {
	Login struct {
		Result string `json:"result"`
	} `json:"login"`
}

func main() {
	inputFile := flag.String("input", "", "Input file with titles")
	outputFile := flag.String("output", "", "Output file to save extracts")
	username := flag.String("username", "", "Wikipedia bot username")
	password := flag.String("password", "", "Wikipedia bot password")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" || *username == "" || *password == "" {
		fmt.Println("Usage: go run main.go --input titles.txt --output wiki.txt --username=botname --password=botpass")
		os.Exit(1)
	}

	// Create output directory
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Open input and output files
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

	// Create HTTP client for session
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
	}

	// Perform login
	loginToken, err := getLoginToken(client)
	if err != nil {
		fmt.Printf("Error getting login token: %v\n", err)
		os.Exit(1)
	}

	err = performLogin(client, *username, *password, loginToken)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	// Process titles
	scanner := bufio.NewScanner(inputHandle)
	for scanner.Scan() {
		title := strings.TrimSpace(scanner.Text())
		if title == "" {
			continue
		}

		extract, err := fetchWikipediaExtract(client, title)
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

func getLoginToken(client *http.Client) (string, error) {
	// API endpoint
	apiURL := "https://bn.wikipedia.org/w/api.php"

	// Prepare token request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	// Set query parameters
	q := req.URL.Query()
	q.Add("action", "query")
	q.Add("meta", "tokens")
	q.Add("type", "login")
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse login token
	var tokenResp LoginTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.Query.Tokens.LoginToken, nil
}

func performLogin(client *http.Client, username, password, loginToken string) error {
	// API endpoint
	apiURL := "https://bn.wikipedia.org/w/api.php"

	// Prepare login data
	data := url.Values{}
	data.Set("action", "login")
	data.Set("lgname", username)
	data.Set("lgpassword", password)
	data.Set("lgtoken", loginToken)
	data.Set("format", "json")

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse login response
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return err
	}

	// Check login result
	if loginResp.Login.Result != "Success" {
		return fmt.Errorf("login failed: %s", loginResp.Login.Result)
	}

	return nil
}

// removeNukta replaces nuktas in Bangla text with their corresponding replacements
func removeNukta(banglaText string) string {
	nuktaReplacements := map[string]string{
		"\u09A1\u09BC": "\u09DC", // \u09DC is 2524 in decimal
		"\u09A2\u09BC": "\u09DD", // \u09DD is 2525 in decimal
		"\u09AF\u09BC": "\u09DF", // \u09DF is 2527 in decimal
	}

	for nuktaChar, replacement := range nuktaReplacements {
		banglaText = strings.ReplaceAll(banglaText, nuktaChar, replacement)
	}

	return banglaText
}

func preprocessText(input string) string {
	// Remove nuktas
	input = removeNukta(input)
	// Define the regex for Bangla words
	banglaWordRegex := regexp.MustCompile(`[\x{0980}-\x{09E5}\x{09F0}-\x{09FF}]+`)
	// Find all matches in the input string
	banglaWords := banglaWordRegex.FindAllString(input, -1)
	return strings.Join(banglaWords, " ")

}

func fetchWikipediaExtract(client *http.Client, title string) (string, error) {
	apiURL := fmt.Sprintf("https://bn.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&explaintext&redirects=1&titles=%s", url.QueryEscape(title))

	resp, err := client.Get(apiURL)
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
		extract = preprocessText(page.Extract)
		break
	}

	if extract == "" {
		return "", fmt.Errorf("no extract found for title: %s", title)
	}

	return extract, nil
}
