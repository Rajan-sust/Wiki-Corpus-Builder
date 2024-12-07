package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

// WordFreq represents a word and its frequency
type WordFreq struct {
	word  string
	count int
}

// MinHeap implementation for WordFreq
type MinHeap []WordFreq

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].count < h[j].count }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(WordFreq))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// processChunk processes a chunk of text and returns a map of word frequencies
func processChunk(chunk string) map[string]int {
	freqMap := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(chunk))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if word != "" {
			freqMap[word]++
		}
	}
	return freqMap
}

// mergeFreqMaps merges multiple frequency maps into one
func mergeFreqMaps(maps []map[string]int) map[string]int {
	result := make(map[string]int)
	for _, m := range maps {
		for word, count := range m {
			result[word] += count
		}
	}
	return result
}

func main() {
	// Command line flags
	n := flag.Int("n", 10, "number of top words to find")
	filename := flag.String("file", "", "input file path")
	flag.Parse()

	if *filename == "" {
		fmt.Println("Please provide an input file using -file flag")
		return
	}

	// Open the file
	file, err := os.Open(*filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a buffered reader
	reader := bufio.NewReader(file)

	// Channel for frequency maps from goroutines
	freqChan := make(chan map[string]int)
	var wg sync.WaitGroup

	// Process the file in chunks
	chunkSize := 64 * 1024 * 1024 // 64MB chunks
	buffer := make([]byte, chunkSize)

	// Start worker goroutines
	go func() {
		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				wg.Add(1)
				chunk := string(buffer[:n])
				go func() {
					defer wg.Done()
					freqChan <- processChunk(chunk)
				}()
			}
			if err != nil {
				break
			}
		}
		wg.Wait()
		close(freqChan)
	}()

	// Collect and merge frequency maps
	var freqMaps []map[string]int
	for freqMap := range freqChan {
		freqMaps = append(freqMaps, freqMap)
	}

	// Merge all frequency maps
	finalFreqMap := mergeFreqMaps(freqMaps)

	// Use a min-heap to find top N words
	h := &MinHeap{}
	heap.Init(h)

	for word, count := range finalFreqMap {
		wordFreq := WordFreq{word: word, count: count}
		if h.Len() < *n {
			heap.Push(h, wordFreq)
		} else if count > (*h)[0].count {
			heap.Pop(h)
			heap.Push(h, wordFreq)
		}
	}

	// Print results in reverse order (highest to lowest frequency)
	results := make([]WordFreq, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		results[i] = heap.Pop(h).(WordFreq)
	}

	fmt.Printf("\nTop %d words by frequency:\n", *n)
	for i := 0; i < len(results); i++ {
		fmt.Printf("%s: %d\n", results[i].word, results[i].count)
	}
}
