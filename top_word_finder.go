package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"
)

type WordCount struct {
	word  string
	count int
}

func main() {
	inputFile := "/home/ovishek/NLP_BrainStorming/word2vecbangla/new_db/merged.txt"
	numWorkers := 12 // Number of worker threads
	top_n := 100     // Number of top words to output
	outputFile := fmt.Sprintf("top_words_%d.txt", top_n)

	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Get file size for progress reporting
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file stats: %v\n", err)
		return
	}
	totalSize := fileInfo.Size()

	// Create channels
	lines := make(chan string, 1000)
	results := make(chan map[string]int, numWorkers)
	progress := make(chan int64, 1000)

	var wg sync.WaitGroup

	// Start progress monitor
	go func() {
		var processedBytes int64
		startTime := time.Now()

		for bytes := range progress {
			processedBytes += bytes
			percentage := float64(processedBytes) / float64(totalSize) * 100
			elapsed := time.Since(startTime).Seconds()
			speed := float64(processedBytes) / (1024 * 1024 * elapsed) // MB/s

			fmt.Printf("\rProgress: %.2f%% (%.2f MB/s)", percentage, speed)
		}
		fmt.Println()
	}()

	// Start worker goroutines
	bengaliRegex := regexp.MustCompile(`[\p{Bengali}]+`)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			wordCounts := make(map[string]int)
			
			for line := range lines {
				matches := bengaliRegex.FindAllString(line, -1)
				for _, word := range matches {
					wordCounts[word]++
				}
			}
			
			results <- wordCounts
			fmt.Printf("\nWorker %d completed\n", workerId)
		}(i)
	}

	// Read file and send lines to workers
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // Increase buffer size

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			lines <- line
			progress <- int64(len(line) + 1) // +1 for newline
		}
		close(lines)
	}()

	// Wait for workers and close results channel
	go func() {
		wg.Wait()
		close(results)
		close(progress)
	}()

	// Merge results
	finalCounts := make(map[string]int)
	for workerCounts := range results {
		for word, count := range workerCounts {
			finalCounts[word] += count
		}
	}

	// Convert to slice for sorting
	var wordCounts []WordCount
	for word, count := range finalCounts {
		wordCounts = append(wordCounts, WordCount{word, count})
	}

	// Sort by count (descending) and then by word
	sort.Slice(wordCounts, func(i, j int) bool {
		if wordCounts[i].count != wordCounts[j].count {
			return wordCounts[i].count > wordCounts[j].count
		}
		return wordCounts[i].word < wordCounts[j].word
	})

	// Write top 200 words to output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for i := 0; i < min(top_n, len(wordCounts)); i++ {
		_, err := fmt.Fprintf(writer, "%d %s\n", wordCounts[i].count, wordCounts[i].word)
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
			return
		}
	}
	writer.Flush()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
